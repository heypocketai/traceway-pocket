package controllers

import (
	"github.com/tracewayapp/traceway/backend/app/middleware"
	"github.com/tracewayapp/traceway/backend/app/models"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/repositories"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	traceway "go.tracewayapp.com"
)

type metricQueryController struct{}

type MetricQueryRequest struct {
	Queries         []MetricQueryItem `json:"queries" binding:"required,min=1"`
	From            time.Time         `json:"from" binding:"required"`
	To              time.Time         `json:"to" binding:"required"`
	IntervalMinutes int               `json:"intervalMinutes"`
}

type MetricQueryItem struct {
	Name        string            `json:"name" binding:"required"`
	Aggregation string            `json:"aggregation"`
	TagFilters  map[string]string `json:"tagFilters"`
	GroupBy     string            `json:"groupBy"`
}

type MetricQueryResponse struct {
	Results []MetricQueryResult `json:"results"`
}

type MetricQueryResult struct {
	Name   string                           `json:"name"`
	Series map[string][]models.TimeSeriesPoint `json:"series"`
}

func (c *metricQueryController) Query(ctx *gin.Context) {
	projectId, err := middleware.GetProjectId(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("RequireProjectAccess middleware must be applied: %w", err))
		return
	}

	var req MetricQueryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.IntervalMinutes <= 0 {
		req.IntervalMinutes = calculateIntervalMinutes(req.To.Sub(req.From))
	}

	var results []MetricQueryResult
	for _, q := range req.Queries {
		agg := q.Aggregation
		if agg == "" {
			agg = "avg"
		}

		series, err := repositories.MetricPointRepository.QueryTimeSeries(
			ctx, projectId, q.Name, req.From, req.To,
			req.IntervalMinutes, agg, q.TagFilters, q.GroupBy,
		)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to query metric %s: %w", q.Name, err))
			return
		}

		results = append(results, MetricQueryResult{
			Name:   q.Name,
			Series: series,
		})
	}

	ctx.JSON(http.StatusOK, MetricQueryResponse{Results: results})
}

type DiscoverResponse struct {
	Metrics []models.DiscoveredMetric `json:"metrics"`
}

func (c *metricQueryController) Discover(ctx *gin.Context) {
	projectId, err := middleware.GetProjectId(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("RequireProjectAccess middleware must be applied: %w", err))
		return
	}

	from := time.Now().AddDate(0, 0, -7)
	to := time.Now()

	if fromStr := ctx.Query("from"); fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = parsed
		}
	}
	if toStr := ctx.Query("to"); toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = parsed
		}
	}

	discovered, err := repositories.MetricPointRepository.DiscoverMetrics(ctx, projectId, from, to)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to discover metrics: %w", err))
		return
	}

	registry, regErr := db.ExecuteTransaction(func(tx *sql.Tx) ([]*models.MetricRegistry, error) {
		return repositories.MetricRegistryRepository.FindByProject(tx, projectId)
	})

	if regErr != nil {
		traceway.CaptureException(fmt.Errorf("failed to load metric registry for discover: %w", regErr))
	}
	if regErr == nil {
		regMap := make(map[string]*models.MetricRegistry, len(registry))
		for _, r := range registry {
			regMap[r.Name] = r
		}
		for i := range discovered {
			if reg, ok := regMap[discovered[i].Name]; ok {
				discovered[i].MetricType = reg.MetricType
				discovered[i].Unit = reg.Unit
			}
		}
	}

	ctx.JSON(http.StatusOK, DiscoverResponse{Metrics: discovered})
}

func (c *metricQueryController) DiscoverTags(ctx *gin.Context) {
	projectId, err := middleware.GetProjectId(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("RequireProjectAccess middleware must be applied: %w", err))
		return
	}

	name := ctx.Query("name")
	key := ctx.Query("key")
	if name == "" || key == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name and key query params are required"})
		return
	}

	from := time.Now().AddDate(0, 0, -7)
	to := time.Now()

	values, err := repositories.MetricPointRepository.DiscoverTagValues(ctx, projectId, name, key, from, to)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to discover tag values: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"values": values})
}

func (c *metricQueryController) UpdateRegistry(ctx *gin.Context) {
	projectId, err := middleware.GetProjectId(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("RequireProjectAccess middleware must be applied: %w", err))
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		MetricType  string `json:"metricType"`
		Unit        string `json:"unit"`
		Description string `json:"description"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MetricType != "" && req.MetricType != "gauge" && req.MetricType != "counter" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "metricType must be 'gauge' or 'counter'"})
		return
	}

	updated, err := db.ExecuteTransaction(func(tx *sql.Tx) (*models.MetricRegistry, error) {
		entry, err := repositories.MetricRegistryRepository.FindByProjectAndName(tx, projectId, req.Name)
		if err != nil {
			return nil, err
		}
		if entry == nil {
			return nil, nil
		}
		if req.MetricType != "" {
			entry.MetricType = req.MetricType
		}
		if req.Unit != "" {
			entry.Unit = req.Unit
		}
		entry.Description = req.Description
		if err := repositories.MetricRegistryRepository.Update(tx, entry); err != nil {
			return nil, err
		}
		return entry, nil
	})

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to update metric registry: %w", err))
		return
	}
	if updated == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "metric not found in registry"})
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

var MetricQueryController = metricQueryController{}
