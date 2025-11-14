package models

import (
	"time"
)

type CredentialType string

const (
	CredentialTypeAPIKey      CredentialType = "api_key"
	CredentialTypeOAuth       CredentialType = "oauth"
	CredentialTypeBasicAuth   CredentialType = "basic_auth"
	CredentialTypeBearerToken CredentialType = "bearer_token"
	CredentialTypeCustom      CredentialType = "custom"
)

// Credential stores encrypted user credentials
type Credential struct {
	ID             string         `json:"id" gorm:"primaryKey"`
	Name           string         `json:"name" gorm:"not null"`
	Type           CredentialType `json:"type" gorm:"not null"`
	Description    *string        `json:"description,omitempty"`
	EncryptedValue string         `json:"-" gorm:"column:encrypted_value;not null"`
	DecryptedValue string         `json:"value,omitempty" gorm:"-"`
	Metadata       string         `json:"metadata,omitempty" gorm:"column:metadata"`
	CreatedAt      time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Credential) TableName() string {
	return "credentials"
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
}
