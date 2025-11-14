package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/workflow-orchestration/backend/database"
	"github.com/workflow-orchestration/backend/engine"
	"github.com/workflow-orchestration/backend/models"
)

type Handler struct {
	db     *database.DB
	engine *engine.WorkflowEngine
}

func NewHandler(db *database.DB, engine *engine.WorkflowEngine) *Handler {
	return &Handler{
		db:     db,
		engine: engine,
	}
}

// GetAllWorkflows handles GET /api/workflows
func (h *Handler) GetAllWorkflows(c *gin.Context) {
	workflows, err := h.db.GetAllWorkflows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    workflows,
	})
}

// GetWorkflow handles GET /api/workflows/:id
func (h *Handler) GetWorkflow(c *gin.Context) {
	id := c.Param("id")

	workflow, err := h.db.GetWorkflow(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	if workflow == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   stringPtr("Workflow not found"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    workflow,
	})
}

// CreateWorkflow handles POST /api/workflows
func (h *Handler) CreateWorkflow(c *gin.Context) {
	var req struct {
		Name        string        `json:"name"`
		Description *string       `json:"description"`
		Nodes       []models.Node `json:"nodes"`
		Edges       []models.Edge `json:"edges"`
		Active      bool          `json:"active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	now := time.Now()
	workflow := &models.Workflow{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Nodes:       req.Nodes,
		Edges:       req.Edges,
		Active:      req.Active,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.db.CreateWorkflow(workflow); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    workflow,
	})
}

// UpdateWorkflow handles PUT /api/workflows/:id
func (h *Handler) UpdateWorkflow(c *gin.Context) {
	id := c.Param("id")

	existing, err := h.db.GetWorkflow(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	if existing == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   stringPtr("Workflow not found"),
		})
		return
	}

	var req struct {
		Name        *string       `json:"name"`
		Description *string       `json:"description"`
		Nodes       []models.Node `json:"nodes"`
		Edges       []models.Edge `json:"edges"`
		Active      *bool         `json:"active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	// Update fields
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = req.Description
	}
	if req.Nodes != nil {
		existing.Nodes = req.Nodes
	}
	if req.Edges != nil {
		existing.Edges = req.Edges
	}
	if req.Active != nil {
		existing.Active = *req.Active
	}
	existing.UpdatedAt = time.Now()

	if err := h.db.UpdateWorkflow(existing); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    existing,
	})
}

// DeleteWorkflow handles DELETE /api/workflows/:id
func (h *Handler) DeleteWorkflow(c *gin.Context) {
	id := c.Param("id")

	if err := h.db.DeleteWorkflow(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
	})
}

// ExecuteWorkflow handles POST /api/workflows/:id/execute
func (h *Handler) ExecuteWorkflow(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Context map[string]interface{} `json:"context"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.Context = make(map[string]interface{})
	}

	execution, err := h.engine.ExecuteWorkflow(id, req.Context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    execution,
	})
}

// GetWorkflowExecutions handles GET /api/workflows/:id/executions
func (h *Handler) GetWorkflowExecutions(c *gin.Context) {
	id := c.Param("id")

	executions, err := h.db.GetWorkflowExecutions(id, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    executions,
	})
}

// GetExecutionDetails handles GET /api/workflows/:workflowId/executions/:executionId
func (h *Handler) GetExecutionDetails(c *gin.Context) {
	executionID := c.Param("executionId")

	execution, err := h.db.GetExecution(executionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	if execution == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   stringPtr("Execution not found"),
		})
		return
	}

	logs, err := h.db.GetExecutionLogs(executionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"execution": execution,
			"logs":      logs,
		},
	})
}

func stringPtr(s string) *string {
	return &s
}
