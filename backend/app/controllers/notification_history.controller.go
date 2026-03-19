package controllers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/middleware"
	"github.com/tracewayapp/traceway/backend/app/models"
	"github.com/tracewayapp/traceway/backend/app/repositories"
	traceway "go.tracewayapp.com"
)

type notificationHistoryController struct{}

type NotificationHistorySearchRequest struct {
	Pagination PaginationParams `json:"pagination"`
	Search     string           `json:"search"`
	FromDate   string           `json:"fromDate"`
	ToDate     string           `json:"toDate"`
}

func (ctrl *notificationHistoryController) List(ctx *gin.Context) {
	projectId, err := middleware.GetProjectId(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("RequireProjectAccess middleware must be applied: %w", err))
		return
	}

	var request NotificationHistorySearchRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	page := request.Pagination.Page
	pageSize := request.Pagination.PageSize

	var fromTime, toTime *time.Time
	if request.FromDate != "" {
		t, err := time.Parse(time.RFC3339, request.FromDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fromDate format"})
			return
		}
		fromTime = &t
	}
	if request.ToDate != "" {
		t, err := time.Parse(time.RFC3339, request.ToDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid toDate format"})
			return
		}
		toTime = &t
	}

	type historyResult struct {
		Items []*models.NotificationHistory
		Total int64
	}

	result, err := db.ExecuteTransaction(func(tx *sql.Tx) (historyResult, error) {
		items, total, err := repositories.NotificationHistoryRepository.FindByProject(tx, projectId, page, pageSize, request.Search, fromTime, toTime)
		return historyResult{Items: items, Total: total}, err
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to list notification history: %w", err))
		return
	}

	totalPages := result.Total / int64(pageSize)
	if result.Total%int64(pageSize) != 0 {
		totalPages++
	}

	data := result.Items
	if data == nil {
		data = []*models.NotificationHistory{}
	}

	ctx.JSON(http.StatusOK, PaginatedResponse[*models.NotificationHistory]{
		Data: data,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      result.Total,
			TotalPages: totalPages,
		},
	})
}

var NotificationHistoryController = notificationHistoryController{}
