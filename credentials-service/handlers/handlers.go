package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/workflow-orchestration/credentials-service/crypto"
	"github.com/workflow-orchestration/credentials-service/database"
	"github.com/workflow-orchestration/credentials-service/models"
)

type Handler struct {
	db *database.DB
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// GetAllCredentials handles GET /api/credentials
func (h *Handler) GetAllCredentials(c *gin.Context) {
	credentials, err := h.db.GetAllCredentials()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	// Never return decrypted values in list view
	for i := range credentials {
		credentials[i].DecryptedValue = ""
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    credentials,
	})
}

// GetCredential handles GET /api/credentials/:id
func (h *Handler) GetCredential(c *gin.Context) {
	id := c.Param("id")

	credential, err := h.db.GetCredential(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	if credential == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   stringPtr("Credential not found"),
		})
		return
	}

	// Decrypt value for retrieval
	decryptedValue, err := crypto.Decrypt(credential.EncryptedValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr("Failed to decrypt credential"),
		})
		return
	}
	credential.DecryptedValue = decryptedValue

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    credential,
	})
}

// CreateCredential handles POST /api/credentials
func (h *Handler) CreateCredential(c *gin.Context) {
	var req struct {
		Name        string                `json:"name"`
		Type        models.CredentialType `json:"type"`
		Description *string               `json:"description"`
		Value       string                `json:"value"`
		Metadata    string                `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	// Encrypt the value
	encryptedValue, err := crypto.Encrypt(req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr("Failed to encrypt credential"),
		})
		return
	}

	now := time.Now()
	credential := &models.Credential{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Type:           req.Type,
		Description:    req.Description,
		EncryptedValue: encryptedValue,
		Metadata:       req.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.db.CreateCredential(credential); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	// Don't return encrypted value
	credential.DecryptedValue = ""

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    credential,
	})
}

// UpdateCredential handles PUT /api/credentials/:id
func (h *Handler) UpdateCredential(c *gin.Context) {
	id := c.Param("id")

	existing, err := h.db.GetCredential(id)
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
			Error:   stringPtr("Credential not found"),
		})
		return
	}

	var req struct {
		Name        *string                `json:"name"`
		Type        *models.CredentialType `json:"type"`
		Description *string                `json:"description"`
		Value       *string                `json:"value"`
		Metadata    *string                `json:"metadata"`
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
	if req.Type != nil {
		existing.Type = *req.Type
	}
	if req.Description != nil {
		existing.Description = req.Description
	}
	if req.Value != nil {
		// Re-encrypt the new value
		encryptedValue, err := crypto.Encrypt(*req.Value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   stringPtr("Failed to encrypt credential"),
			})
			return
		}
		existing.EncryptedValue = encryptedValue
	}
	if req.Metadata != nil {
		existing.Metadata = *req.Metadata
	}
	existing.UpdatedAt = time.Now()

	if err := h.db.UpdateCredential(existing); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   stringPtr(err.Error()),
		})
		return
	}

	// Don't return encrypted value
	existing.DecryptedValue = ""

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    existing,
	})
}

// DeleteCredential handles DELETE /api/credentials/:id
func (h *Handler) DeleteCredential(c *gin.Context) {
	id := c.Param("id")

	if err := h.db.DeleteCredential(id); err != nil {
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
