package clientcontrollers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/hooks"
	"github.com/tracewayapp/traceway/backend/app/middleware"
	"github.com/tracewayapp/traceway/backend/app/models"
	"github.com/tracewayapp/traceway/backend/app/models/clientmodels"
	"github.com/tracewayapp/traceway/backend/app/monitoring"
	"github.com/tracewayapp/traceway/backend/app/repositories"
	"github.com/tracewayapp/traceway/backend/app/services"
	"github.com/tracewayapp/traceway/backend/app/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	traceway "go.tracewayapp.com"
)

type clientController struct{}

// isEmptyRaw reports whether a json.RawMessage carries no meaningful payload —
// nil, blank, `null`, `[]`, or `{}` all count as empty. Used to drop session
// recordings that would otherwise just be wasted S3 writes.
func isEmptyRaw(r json.RawMessage) bool {
	if len(r) == 0 {
		return true
	}
	trimmed := bytes.TrimSpace(r)
	return bytes.Equal(trimmed, []byte("null")) ||
		bytes.Equal(trimmed, []byte("[]")) ||
		bytes.Equal(trimmed, []byte("{}"))
}

type ReportRequest struct {
	CollectionFrames []*clientmodels.CollectionFrame `json:"collectionFrames"`
	AppVersion       string                          `json:"appVersion"`
	ServerName       string                          `json:"serverName"`
}

func (e clientController) Report(c *gin.Context) {
	projectId, err := middleware.GetProjectId(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("UseClientAuth middleware must be applied: %w", err))
		return
	}

	if project, exists := c.Get(middleware.ProjectContextKey); exists {
		if p, ok := project.(*models.Project); ok && p.OrganizationId != nil {
			if !hooks.CanReport(*p.OrganizationId) {
				monitoring.RecordRateLimited(*p.OrganizationId)
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
		}
	}

	parseSpan := traceway.StartSpan(c, "report.parse_body")
	var request ReportRequest
	if err := c.ShouldBindBodyWithJSON(&request); err != nil {
		parseSpan.End()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	parseSpan.End()

	convertStart := time.Now()

	endpointsToInsert := []models.Endpoint{}
	tasksToInsert := []models.Task{}
	exceptionStackTraceToInsert := []models.ExceptionStackTrace{}
	metricPointsToInsert := []models.MetricPoint{}
	spansToInsert := []models.Span{}

	type recordingWork struct {
		Id          uuid.UUID
		ProjectId   uuid.UUID
		ExceptionId uuid.UUID
		// Body is the marshaled JSON of the entire ClientSessionRecording
		// sub-document — events + logs + actions + startedAt/endedAt — exactly
		// as it lands in S3. App console logs in `logs` are intentionally not
		// inserted into the OTel logs ClickHouse table; they live only here.
		Body       []byte
		RecordedAt time.Time
	}
	var recordingsWork []recordingWork

	// Map frontend sessionRecordingId → backend-generated exception UUID
	recordingIdToExceptionId := map[string]uuid.UUID{}

	convertSpan := traceway.StartSpan(c, "report.convert_frames")
	for _, cf := range request.CollectionFrames {
		for _, ct := range cf.Traces {
			if ct.IsTask {
				t := ct.ToTask(request.AppVersion, request.ServerName)
				t.ProjectId = projectId
				tasksToInsert = append(tasksToInsert, t)
			} else {
				e := ct.ToEndpoint(request.AppVersion, request.ServerName)
				e.ProjectId = projectId
				if e.StatusCode == 404 {
					e.Endpoint = "UNMATCHED"
				}
				endpointsToInsert = append(endpointsToInsert, e)
			}

			for _, cs := range ct.Spans {
				span := cs.ToSpan(ct.ParsedId())
				span.ProjectId = projectId
				spansToInsert = append(spansToInsert, span)
			}
		}
		projectAsAny, projectExists := c.Get(middleware.ProjectContextKey)
		var project *models.Project
		if projectExists {
			if p, ok := projectAsAny.(*models.Project); ok {
				project = p
			}
		}

		var sourceMaps *[]*models.SourceMap
		if project != nil && isJsFramework(project.Framework) {
			loadSpan := traceway.StartSpan(c, "report.load_source_maps")
			sourceMapsLoaded, err := db.ExecuteTransaction(func(tx *sql.Tx) ([]*models.SourceMap, error) {
				if request.AppVersion != "" {
					return repositories.SourceMapRepository.FindByProjectAndVersion(tx, projectId, request.AppVersion)
				}
				return repositories.SourceMapRepository.FindLatestByProject(tx, projectId)
			})
			loadSpan.End()
			if err == nil && len(sourceMapsLoaded) > 0 {
				sourceMaps = &sourceMapsLoaded
			}
		}

		resolveSpan := traceway.StartSpan(c, "report.resolve_stack_traces")
		for _, cst := range cf.StackTraces {
			resolvedStackTrace := cst.StackTrace
			if sourceMaps != nil {
				resolvedStackTrace = services.ResolveStackTrace(c, projectId, cst.StackTrace, *sourceMaps)
			}
			est := cst.ToExceptionStackTrace(ComputeExceptionHash(resolvedStackTrace, cst.IsMessage), request.AppVersion, request.ServerName)
			est.StackTrace = resolvedStackTrace
			est.Id = uuid.New()
			est.ProjectId = projectId
			if cst.SessionRecordingId != nil {
				recordingIdToExceptionId[*cst.SessionRecordingId] = est.Id
			}
			exceptionStackTraceToInsert = append(exceptionStackTraceToInsert, est)
		}
		resolveSpan.End()

		for _, cm := range cf.Metrics {
			mp := cm.ToMetricPoint(request.ServerName)
			mp.ProjectId = projectId
			metricPointsToInsert = append(metricPointsToInsert, mp)
		}

		for _, sr := range cf.SessionRecordings {
			exceptionId, ok := recordingIdToExceptionId[sr.ExceptionId]
			if !ok {
				continue
			}
			// Skip recordings that carry no payload at all. The SDK shouldn't
			// be sending these, but be defensive — an empty recording would
			// just be a wasted S3 write and an empty session_recordings row.
			if isEmptyRaw(sr.Events) && isEmptyRaw(sr.Logs) && isEmptyRaw(sr.Actions) {
				continue
			}
			body, err := json.Marshal(sr)
			if err != nil {
				traceway.CaptureException(traceway.NewStackTraceErrorf("failed to marshal session recording: %w", err))
				continue
			}
			recordingsWork = append(recordingsWork, recordingWork{
				Id:          uuid.New(),
				ProjectId:   projectId,
				ExceptionId: exceptionId,
				Body:        body,
				RecordedAt:  time.Now().UTC(),
			})
		}
	}
	convertSpan.End()

	convertMs := float64(time.Since(convertStart).Microseconds()) / 1000.0
	insertStart := time.Now()

	if len(endpointsToInsert) > 0 {
		insertSpan := traceway.StartSpan(c, "report.insert.endpoints")
		err := repositories.EndpointRepository.InsertAsync(c, endpointsToInsert)
		insertSpan.End()
		if err != nil {
			c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting endpointsToInsert: %w", err))
			return
		}
	}

	if len(tasksToInsert) > 0 {
		insertSpan := traceway.StartSpan(c, "report.insert.tasks")
		err := repositories.TaskRepository.InsertAsync(c, tasksToInsert)
		insertSpan.End()
		if err != nil {
			c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting tasksToInsert: %w", err))
			return
		}
	}

	exceptionInsertSpan := traceway.StartSpan(c, "report.insert.exceptions")
	err = repositories.ExceptionStackTraceRepository.InsertAsync(c, exceptionStackTraceToInsert)
	exceptionInsertSpan.End()

	if err != nil {
		c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting exceptionStackTraceToInsert: %w", err))
		return
	}

	if len(metricPointsToInsert) > 0 {
		insertSpan := traceway.StartSpan(c, "report.insert.metric_points")
		err := repositories.MetricPointRepository.InsertAsync(c, metricPointsToInsert)
		insertSpan.End()
		if err != nil {
			c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting metricPointsToInsert: %w", err))
			return
		}

		metricNames := services.CollectUniqueMetricNames(metricPointsToInsert)
		go services.AutoRegisterMetrics(projectId, metricNames)
	}

	spanInsertSpan := traceway.StartSpan(c, "report.insert.spans")
	err = repositories.SpanRepository.InsertAsync(c, spansToInsert)
	spanInsertSpan.End()

	if err != nil {
		c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting spansToInsert: %w", err))
		return
	}

	insertMs := float64(time.Since(insertStart).Microseconds()) / 1000.0
	totalSize := len(endpointsToInsert) + len(tasksToInsert) + len(spansToInsert) + len(exceptionStackTraceToInsert) + len(metricPointsToInsert)
	monitoring.RecordIngestBatch(monitoring.SignalNative, "report", convertMs, insertMs, totalSize)

	var exceptionHashes []string
	for _, est := range exceptionStackTraceToInsert {
		exceptionHashes = append(exceptionHashes, est.ExceptionHash)
	}

	if project, exists := c.Get(middleware.ProjectContextKey); exists {
		if p, ok := project.(*models.Project); ok && p.OrganizationId != nil {
			hooks.BroadcastReport(hooks.ReportEvent{
				OrganizationId:  *p.OrganizationId,
				ProjectId:       projectId,
				EndpointCount:   len(endpointsToInsert),
				ErrorCount:      len(exceptionStackTraceToInsert),
				TaskCount:       len(tasksToInsert),
				RecordingCount:  len(recordingsWork),
				ExceptionHashes: exceptionHashes,
			})
		}
	}

	if len(recordingsWork) > 0 {
		work := recordingsWork
		go func() {
			var successful []models.SessionRecording
			for _, rw := range work {
				key := fmt.Sprintf("recordings/%s/%s.json", rw.ProjectId, rw.ExceptionId)
				if err := storage.Store.Write(context.Background(), key, rw.Body); err != nil {
					traceway.CaptureException(traceway.NewStackTraceErrorf("failed to write session recording (key=%s): %w", key, err))
					continue
				}
				successful = append(successful, models.SessionRecording{
					Id:          rw.Id,
					ProjectId:   rw.ProjectId,
					ExceptionId: rw.ExceptionId,
					FilePath:    key,
					RecordedAt:  rw.RecordedAt,
				})
			}
			if len(successful) > 0 {
				if err := repositories.SessionRecordingRepository.InsertAsync(context.Background(), successful); err != nil {
					traceway.CaptureException(traceway.NewStackTraceErrorf("failed to batch insert %d session recording refs: %w", len(successful), err))
				}
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{})
}

var (
	errorMessageRe = regexp.MustCompile(`(?m)^(\*?[\w.]+):\s*.+`)
	absolutePathRe = regexp.MustCompile(`/[^\s:]+/([^/\s:]+:\d+)`)
	versionRe      = regexp.MustCompile(`@v[\d.]+`)
	hexRe          = regexp.MustCompile(`0x[0-9a-fA-F]+`)
	uuidRe         = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	largeNumberRe  = regexp.MustCompile(`(^|[^:\d])(\d{5,})($|[^\d])`)
	emailRe        = regexp.MustCompile(`[\w.\-]+@[\w.\-]+\.\w+`)
	ipRe           = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}(:\d+)?`)
	goroutineRe    = regexp.MustCompile(`goroutine \d+`)
	spacesRe       = regexp.MustCompile(`[ \t]+`)
	newlinesRe     = regexp.MustCompile(`\n+`)
)

func ComputeExceptionHash(stackTrace string, isMessage bool) string {
	normalized := stackTrace

	if !isMessage {
		normalized = errorMessageRe.ReplaceAllString(normalized, "$1")
		normalized = absolutePathRe.ReplaceAllString(normalized, "$1")
		normalized = versionRe.ReplaceAllString(normalized, "")
		normalized = hexRe.ReplaceAllString(normalized, "<hex>")
		normalized = uuidRe.ReplaceAllString(normalized, "<uuid>")
		normalized = largeNumberRe.ReplaceAllString(normalized, "${1}<id>${3}")
		normalized = emailRe.ReplaceAllString(normalized, "<email>")
		normalized = ipRe.ReplaceAllString(normalized, "<ip>")
		normalized = goroutineRe.ReplaceAllString(normalized, "goroutine <n>")
		normalized = spacesRe.ReplaceAllString(normalized, " ")
		normalized = newlinesRe.ReplaceAllString(normalized, "\n")
	}

	normalized = strings.TrimSpace(normalized)
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])[:16]
}

var jsFrameworks = map[string]bool{
	"react":   true,
	"svelte":  true,
	"vuejs":   true,
	"nextjs":  true,
	"nestjs":  true,
	"express": true,
	"remix":   true,
}

func isJsFramework(framework string) bool {
	return jsFrameworks[framework]
}

var ClientController = clientController{}
