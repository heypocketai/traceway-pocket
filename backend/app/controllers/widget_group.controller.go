package controllers

import (
	"github.com/tracewayapp/traceway/backend/app/middleware"
	"github.com/tracewayapp/traceway/backend/app/models"
	"github.com/tracewayapp/traceway/backend/app/repositories"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	traceway "go.tracewayapp.com"
)

type widgetGroupController struct{}

func (c *widgetGroupController) List(ctx *gin.Context) {
	projectId, err := middleware.GetProjectId(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("RequireProjectAccess middleware must be applied: %w", err))
		return
	}

	tx := middleware.GetTx(ctx)

	list, err := repositories.WidgetGroupRepository.FindByProject(tx, projectId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to list widget groups: %w", err))
		return
	}
	if len(list) == 0 {
		project, err := repositories.ProjectRepository.FindById(tx, projectId)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to find project: %w", err))
			return
		}
		if project == nil || project.Framework != "opentelemetry" {
			if err := ensureDefaultWidgetGroups(tx, projectId); err != nil {
				ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to create default widget groups: %w", err))
				return
			}
			list, err = repositories.WidgetGroupRepository.FindByProject(tx, projectId)
			if err != nil {
				ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to list widget groups: %w", err))
				return
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"widgetGroups": list})
}

type CreateWidgetGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *widgetGroupController) Create(ctx *gin.Context) {
	projectId, err := middleware.GetProjectId(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("RequireProjectAccess middleware must be applied: %w", err))
		return
	}

	var req CreateWidgetGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Name is required."})
		return
	}
	if len(req.Name) > 12 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Name must be 12 characters or fewer."})
		return
	}

	tx := middleware.GetTx(ctx)

	existing, err := repositories.WidgetGroupRepository.FindByProject(tx, projectId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to check duplicate widget group name: %w", err))
		return
	}
	for _, g := range existing {
		if strings.EqualFold(g.Name, req.Name) {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "A group with this name already exists."})
			return
		}
	}

	userId := middleware.GetUserId(ctx)
	var createdBy *int
	if userId > 0 {
		createdBy = &userId
	}

	g := &models.WidgetGroup{
		ProjectId:   projectId,
		Name:        req.Name,
		Description: req.Description,
		IsDefault:   false,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	id, err := repositories.WidgetGroupRepository.Create(tx, g)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to create widget group: %w", err))
		return
	}
	g.Id = id

	ctx.JSON(http.StatusCreated, g)
}

func (c *widgetGroupController) GetWithWidgets(ctx *gin.Context) {
	projectId, err := middleware.GetProjectId(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("RequireProjectAccess middleware must be applied: %w", err))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid widget group id"})
		return
	}

	tx := middleware.GetTx(ctx)

	group, err := repositories.WidgetGroupRepository.FindById(tx, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to get widget group: %w", err))
		return
	}
	if group == nil || group.ProjectId != projectId {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "widget group not found"})
		return
	}

	widgets, err := repositories.WidgetGroupRepository.FindWidgetsByGroup(tx, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to get widget group widgets: %w", err))
		return
	}

	widgetSlice := []models.WidgetGroupWidget{}
	for _, w := range widgets {
		widgetSlice = append(widgetSlice, *w)
	}

	ctx.JSON(http.StatusOK, &models.WidgetGroupWithWidgets{
		WidgetGroup: *group,
		Widgets:     widgetSlice,
	})
}

func (c *widgetGroupController) Update(ctx *gin.Context) {
	projectId, err := middleware.GetProjectId(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("RequireProjectAccess middleware must be applied: %w", err))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid widget group id"})
		return
	}

	var req CreateWidgetGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Name is required."})
		return
	}
	if len(req.Name) > 12 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Name must be 12 characters or fewer."})
		return
	}

	tx := middleware.GetTx(ctx)

	existing, err := repositories.WidgetGroupRepository.FindById(tx, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to update widget group: %w", err))
		return
	}
	if existing == nil || existing.ProjectId != projectId {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "widget group not found"})
		return
	}

	allGroups, err := repositories.WidgetGroupRepository.FindByProject(tx, projectId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to check duplicate widget group name: %w", err))
		return
	}
	for _, g := range allGroups {
		if g.Id != id && strings.EqualFold(g.Name, req.Name) {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "A group with this name already exists."})
			return
		}
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.UpdatedAt = time.Now().UTC()

	if err := repositories.WidgetGroupRepository.Update(tx, existing); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to update widget group: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, existing)
}

func (c *widgetGroupController) Delete(ctx *gin.Context) {
	projectId, err := middleware.GetProjectId(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("RequireProjectAccess middleware must be applied: %w", err))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid widget group id"})
		return
	}

	tx := middleware.GetTx(ctx)

	existing, err := repositories.WidgetGroupRepository.FindById(tx, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to delete widget group: %w", err))
		return
	}
	if existing == nil || existing.ProjectId != projectId {
		ctx.JSON(http.StatusOK, gin.H{"deleted": true})
		return
	}

	if existing.IsDefault {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Default groups cannot be deleted."})
		return
	}

	if err := repositories.WidgetGroupRepository.Delete(tx, id); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, traceway.NewStackTraceErrorf("failed to delete widget group: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}

func ensureDefaultWidgetGroups(tx *sql.Tx, projectId uuid.UUID) error {
	project, err := repositories.ProjectRepository.FindById(tx, projectId)
	if err != nil {
		return err
	}

	jsFrameworks := map[string]bool{
		"react": true, "svelte": true, "vuejs": true,
		"nextjs": true, "nestjs": true, "express": true, "remix": true,
	}
	isJS := project != nil && jsFrameworks[project.Framework]

	type widgetDef struct {
		title string
		name  string
	}
	type groupDef struct {
		name    string
		widgets []widgetDef
	}

	var groupDefs []groupDef

	if !isJS {
		groupDefs = append(groupDefs, groupDef{
			name: "Application",
			widgets: []widgetDef{
				{"Go Routines", "go.go_routines"},
				{"Heap Objects", "go.heap_objects"},
				{"GC Cycles", "go.num_gc"},
				{"GC Pause", "go.gc_pause"},
			},
		})
	}

	groupDefs = append(groupDefs,
		groupDef{
			name: "Stats",
			widgets: []widgetDef{
				{"Memory Usage", "mem.used"},
				{"Memory Used %", "mem.used_pcnt"},
				{"Total Memory", "mem.total"},
			},
		},
		groupDef{
			name: "CPU / Mem",
			widgets: []widgetDef{
				{"CPU Usage", "cpu.used_pcnt"},
				{"Memory Usage", "mem.used"},
			},
		},
	)

	now := time.Now().UTC()
	for _, gd := range groupDefs {
		g := &models.WidgetGroup{
			ProjectId: projectId,
			Name:      gd.name,
			IsDefault: true,
			CreatedAt: now,
			UpdatedAt: now,
		}
		groupId, err := repositories.WidgetGroupRepository.Create(tx, g)
		if err != nil {
			return err
		}

		for i, w := range gd.widgets {
			config := json.RawMessage(`{"sources":[{"type":"metric","name":"` + w.name + `","aggregation":"avg"}]}`)
			widget := &models.WidgetGroupWidget{
				WidgetGroupId: groupId,
				Title:         w.title,
				WidgetType:    "line_chart",
				Config:        config,
				Position:      i,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if _, err := repositories.WidgetGroupRepository.CreateWidget(tx, widget); err != nil {
				return err
			}
		}
	}

	return nil
}

var WidgetGroupController = widgetGroupController{}
