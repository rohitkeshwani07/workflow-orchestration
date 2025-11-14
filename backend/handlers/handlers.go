package handlers

import (
	"io"
	"net/http"
	"os"
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

// GetExecutionDetails handles GET /api/workflows/:id/executions/:executionId
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

// Credentials proxy handlers - forward to credentials service
func (h *Handler) ProxyToCredentialsService(c *gin.Context) {
	credentialsServiceURL := os.Getenv("CREDENTIALS_SERVICE_URL")
	if credentialsServiceURL == "" {
		credentialsServiceURL = "http://localhost:3002"
	}

	// Build target URL
	targetURL := credentialsServiceURL + c.Request.URL.Path

	// Create request to credentials service
	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr("Failed to create proxy request"),
		})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.APIResponse{
			Success: false,
			Error:   stringPtr("Failed to reach credentials service"),
		})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	// Copy status code
	c.Status(resp.StatusCode)

	// Copy response body
	io.Copy(c.Writer, resp.Body)
}

// GetAllWorkflowTemplates handles GET /api/templates
func (h *Handler) GetAllWorkflowTemplates(c *gin.Context) {
	category := c.Query("category")

	var templates []models.WorkflowTemplate
	var err error

	if category != "" {
		templates, err = h.db.GetWorkflowTemplatesByCategory(category)
	} else {
		templates, err = h.db.GetAllWorkflowTemplates()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    templates,
	})
}

// GetWorkflowTemplate handles GET /api/templates/:id
func (h *Handler) GetWorkflowTemplate(c *gin.Context) {
	id := c.Param("id")

	template, err := h.db.GetWorkflowTemplate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	if template == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   stringPtr("Template not found"),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    template,
	})
}

// CreateWorkflowFromTemplate handles POST /api/templates/:id/create-workflow
func (h *Handler) CreateWorkflowFromTemplate(c *gin.Context) {
	templateID := c.Param("id")

	template, err := h.db.GetWorkflowTemplate(templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	if template == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   stringPtr("Template not found"),
		})
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to template name if not provided
		req.Name = &template.Name
	}

	// Create workflow from template
	now := time.Now()
	workflowName := template.Name
	if req.Name != nil {
		workflowName = *req.Name
	}

	workflowDesc := template.Description
	if req.Description != nil {
		workflowDesc = req.Description
	}

	workflow := &models.Workflow{
		ID:          uuid.New().String(),
		Name:        workflowName,
		Description: workflowDesc,
		Nodes:       template.Nodes,
		Edges:       template.Edges,
		Active:      false,
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

// CreateWorkflowTemplate handles POST /api/templates
func (h *Handler) CreateWorkflowTemplate(c *gin.Context) {
	var req struct {
		Name        string        `json:"name"`
		Description *string       `json:"description"`
		Category    string        `json:"category"`
		Tags        []string      `json:"tags"`
		Nodes       []models.Node `json:"nodes"`
		Edges       []models.Edge `json:"edges"`
		Thumbnail   *string       `json:"thumbnail"`
		Featured    bool          `json:"featured"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	now := time.Now()
	template := &models.WorkflowTemplate{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Nodes:       req.Nodes,
		Edges:       req.Edges,
		Thumbnail:   req.Thumbnail,
		Featured:    req.Featured,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.db.CreateWorkflowTemplate(template); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    template,
	})
}

// UpdateWorkflowTemplate handles PUT /api/templates/:id
func (h *Handler) UpdateWorkflowTemplate(c *gin.Context) {
	id := c.Param("id")

	existing, err := h.db.GetWorkflowTemplate(id)
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
			Error:   stringPtr("Template not found"),
		})
		return
	}

	var req struct {
		Name        *string       `json:"name"`
		Description *string       `json:"description"`
		Category    *string       `json:"category"`
		Tags        []string      `json:"tags"`
		Nodes       []models.Node `json:"nodes"`
		Edges       []models.Edge `json:"edges"`
		Thumbnail   *string       `json:"thumbnail"`
		Featured    *bool         `json:"featured"`
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
	if req.Category != nil {
		existing.Category = *req.Category
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}
	if req.Nodes != nil {
		existing.Nodes = req.Nodes
	}
	if req.Edges != nil {
		existing.Edges = req.Edges
	}
	if req.Thumbnail != nil {
		existing.Thumbnail = req.Thumbnail
	}
	if req.Featured != nil {
		existing.Featured = *req.Featured
	}
	existing.UpdatedAt = time.Now()

	if err := h.db.UpdateWorkflowTemplate(existing); err != nil {
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

// DeleteWorkflowTemplate handles DELETE /api/templates/:id
func (h *Handler) DeleteWorkflowTemplate(c *gin.Context) {
	id := c.Param("id")

	if err := h.db.DeleteWorkflowTemplate(id); err != nil {
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

func stringPtr(s string) *string {
	return &s
}
