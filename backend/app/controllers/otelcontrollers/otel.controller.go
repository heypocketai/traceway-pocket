package otelcontrollers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tracewayapp/traceway/backend/app/hooks"
	"github.com/tracewayapp/traceway/backend/app/middleware"
	"github.com/tracewayapp/traceway/backend/app/models"
	"github.com/tracewayapp/traceway/backend/app/repositories"
	"github.com/tracewayapp/traceway/backend/app/services"
	"github.com/tracewayapp/traceway/backend/app/storage"
	traceway "go.tracewayapp.com"
	"google.golang.org/protobuf/encoding/protojson"
)

type otelController struct{}

var OtelController = otelController{}

func (o otelController) ExportTraces(c *gin.Context) {
	fmt.Println("A00")
	projectId, err := middleware.GetProjectId(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("UseClientAuth middleware must be applied: %w", err))
		return
	}
	fmt.Println("A0")
	if project, exists := c.Get(middleware.ProjectContextKey); exists {
		if p, ok := project.(*models.Project); ok && p.OrganizationId != nil {
			if !hooks.CanReport(*p.OrganizationId) {
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
		}
	}
	fmt.Println("A1")

	req, err := decodeTraceRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("WTF??")
	if jsonBytes, err := protojson.Marshal(req); err == nil {
		log.Printf("[OTEL TRACES] Payload: %s", string(jsonBytes))
	}

	endpoints, tasks, spans, exceptions, aiTraces, aiConversations := convertTraces(projectId, req)

	if len(endpoints) > 0 {
		if err := repositories.EndpointRepository.InsertAsync(c, endpoints); err != nil {
			c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting OTEL endpoints: %w", err))
			return
		}
	}

	if len(tasks) > 0 {
		if err := repositories.TaskRepository.InsertAsync(c, tasks); err != nil {
			c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting OTEL tasks: %w", err))
			return
		}
	}

	if err := repositories.ExceptionStackTraceRepository.InsertAsync(c, exceptions); err != nil {
		c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting OTEL exceptions: %w", err))
		return
	}

	if err := repositories.SpanRepository.InsertAsync(c, spans); err != nil {
		c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting OTEL spans: %w", err))
		return
	}

	if len(aiTraces) > 0 {
		if err := repositories.AiTraceRepository.InsertAsync(c, aiTraces); err != nil {
			c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting OTEL ai traces: %w", err))
			return
		}

		if len(aiConversations) > 0 {
			convs := aiConversations
			go func() {
				for _, conv := range convs {
					if err := storage.Store.Write(context.Background(), conv.StorageKey, conv.Content); err != nil {
						traceway.CaptureException(fmt.Errorf("failed to write AI trace conversation (key=%s): %w", conv.StorageKey, err))
					}
				}
			}()
		}
	}

	var exceptionHashes []string
	for _, ex := range exceptions {
		exceptionHashes = append(exceptionHashes, ex.ExceptionHash)
	}

	if project, exists := c.Get(middleware.ProjectContextKey); exists {
		if p, ok := project.(*models.Project); ok && p.OrganizationId != nil {
			hooks.BroadcastReport(hooks.ReportEvent{
				OrganizationId:  *p.OrganizationId,
				ProjectId:       projectId,
				EndpointCount:   len(endpoints),
				ErrorCount:      len(exceptions),
				TaskCount:       len(tasks),
				ExceptionHashes: exceptionHashes,
			})
		}
	}

	writeTraceResponse(c)
}

func (o otelController) ExportMetrics(c *gin.Context) {
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

	req, err := decodeMetricsRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := convertMetricPoints(projectId, req)
	if len(result.Points) > 0 {
		if err := repositories.MetricPointRepository.InsertAsync(c, result.Points); err != nil {
			c.AbortWithError(500, traceway.NewStackTraceErrorf("error inserting OTEL metric points: %w", err))
			return
		}

		if len(result.Entries) > 0 {
			go services.AutoRegisterMetricsWithUnits(projectId, result.Entries)
		}
	}

	writeMetricsResponse(c)
}
