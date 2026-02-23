package clientcontrollers

import (
	"github.com/tracewayapp/traceway/backend/app/hooks"
	"github.com/tracewayapp/traceway/backend/app/middleware"
	"github.com/tracewayapp/traceway/backend/app/models"
	"github.com/tracewayapp/traceway/backend/app/models/clientmodels"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/repositories"
	"github.com/tracewayapp/traceway/backend/app/services"
	"github.com/tracewayapp/traceway/backend/app/storage"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	traceway "go.tracewayapp.com"
)

type clientController struct{}

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
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
		}
	}

	var request ReportRequest
	if err := c.ShouldBindBodyWithJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	endpointsToInsert := []models.Endpoint{}
	tasksToInsert := []models.Task{}
	exceptionStackTraceToInsert := []models.ExceptionStackTrace{}
	metricPointsToInsert := []models.MetricPoint{}
	spansToInsert := []models.Span{}

	type recordingWork struct {
		Id          uuid.UUID
		ProjectId   uuid.UUID
		ExceptionId uuid.UUID
		Events      []byte
		RecordedAt  time.Time
	}
	var recordingsWork []recordingWork

	// Map frontend sessionRecordingId → backend-generated exception UUID
	recordingIdToExceptionId := map[string]uuid.UUID{}

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
			sourceMapsLoaded, err := db.ExecuteTransaction(func(tx *sql.Tx) ([]*models.SourceMap, error) {
				if request.AppVersion != "" {
					return repositories.SourceMapRepository.FindByProjectAndVersion(tx, projectId, request.AppVersion)
				}
				return repositories.SourceMapRepository.FindLatestByProject(tx, projectId)
			})
			if err == nil && len(sourceMapsLoaded) > 0 {
				sourceMaps = &sourceMapsLoaded
			}
		}

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
			recordingsWork = append(recordingsWork, recordingWork{
				Id:          uuid.New(),
				ProjectId:   projectId,
				ExceptionId: exceptionId,
				Events:      sr.Events,
				RecordedAt:  time.Now().UTC(),
			})
		}
	}

	if len(endpointsToInsert) > 0 {
		err := repositories.EndpointRepository.InsertAsync(c, endpointsToInsert)
		if err != nil {
			c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting endpointsToInsert: %w", err))
			return
		}
	}

	if len(tasksToInsert) > 0 {
		err := repositories.TaskRepository.InsertAsync(c, tasksToInsert)
		if err != nil {
			c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting tasksToInsert: %w", err))
			return
		}
	}

	err = repositories.ExceptionStackTraceRepository.InsertAsync(c, exceptionStackTraceToInsert)

	if err != nil {
		c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting exceptionStackTraceToInsert: %w", err))
		return
	}

	if len(metricPointsToInsert) > 0 {
		if err := repositories.MetricPointRepository.InsertAsync(c, metricPointsToInsert); err != nil {
			c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting metricPointsToInsert: %w", err))
			return
		}

		metricNames := collectUniqueMetricNames(metricPointsToInsert)
		go autoRegisterMetrics(projectId, metricNames)
	}

	err = repositories.SpanRepository.InsertAsync(c, spansToInsert)

	if err != nil {
		c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting spansToInsert: %w", err))
		return
	}

	if project, exists := c.Get(middleware.ProjectContextKey); exists {
		if p, ok := project.(*models.Project); ok && p.OrganizationId != nil {
			hooks.BroadcastReport(hooks.ReportEvent{
				OrganizationId: *p.OrganizationId,
				EndpointCount:  len(endpointsToInsert),
				ErrorCount:     len(exceptionStackTraceToInsert),
				TaskCount:      len(tasksToInsert),
				RecordingCount: len(recordingsWork),
			})
		}
	}

	if len(recordingsWork) > 0 {
		work := recordingsWork
		go func() {
			var successful []models.SessionRecording
			for _, rw := range work {
				key := fmt.Sprintf("recordings/%s/%s.json", rw.ProjectId, rw.ExceptionId)
				if err := storage.Store.Write(context.Background(), key, rw.Events); err != nil {
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

func collectUniqueMetricNames(points []models.MetricPoint) []string {
	seen := make(map[string]struct{}, len(points))
	var names []string
	for _, p := range points {
		if _, ok := seen[p.Name]; !ok {
			seen[p.Name] = struct{}{}
			names = append(names, p.Name)
		}
	}
	return names
}

func autoRegisterMetrics(projectId uuid.UUID, names []string) {
	_, err := db.ExecuteTransaction(func(tx *sql.Tx) (struct{}, error) {
		return struct{}{}, repositories.MetricRegistryRepository.EnsureRegistered(tx, projectId, names)
	})
	if err != nil {
		traceway.CaptureException(fmt.Errorf("failed to auto-register metrics: %w", err))
	}
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
