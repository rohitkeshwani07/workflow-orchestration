package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type NodeType string

const (
	NodeTypeChatTrigger  NodeType = "chat_trigger"
	NodeTypeAIAgent      NodeType = "ai_agent"
	NodeTypeHTTPRequest  NodeType = "http_request"
	NodeTypeTransform    NodeType = "transform"
	NodeTypeCondition    NodeType = "condition"
	NodeTypeCode         NodeType = "code"
	NodeTypeDelay        NodeType = "delay"
	NodeTypeResponse     NodeType = "response"
)

// JSON type for GORM
type JSON json.RawMessage

// Scan implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = JSON("null")
		return nil
	}

	// Handle both []byte and string types
	switch v := value.(type) {
	case []byte:
		*j = JSON(v)
		return nil
	case string:
		*j = JSON(v)
		return nil
	default:
		return errors.New("type assertion failed: expected []byte or string")
	}
}

// Value implements driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// MarshalJSON for json encoding
func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

// UnmarshalJSON for json decoding
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSON: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Node struct {
	ID       string                 `json:"id"`
	Type     NodeType               `json:"type"`
	Position Position               `json:"position" gorm:"embedded;embeddedPrefix:position_"`
	Data     map[string]interface{} `json:"data" gorm:"-"`
	DataJSON JSON                   `json:"-" gorm:"column:data"`
}

// BeforeSave hook to convert Data to JSON
func (n *Node) BeforeSave(tx *gorm.DB) error {
	// Always set DataJSON to avoid NULL constraint violations
	if n.Data == nil {
		n.DataJSON = JSON("{}")
	} else {
		data, err := json.Marshal(n.Data)
		if err != nil {
			return err
		}
		n.DataJSON = JSON(data)
	}
	return nil
}

// AfterFind hook to convert JSON to Data
func (n *Node) AfterFind(tx *gorm.DB) error {
	if len(n.DataJSON) > 0 && string(n.DataJSON) != "null" {
		return json.Unmarshal(n.DataJSON, &n.Data)
	}
	return nil
}

type Edge struct {
	ID           string  `json:"id"`
	Source       string  `json:"source"`
	Target       string  `json:"target"`
	SourceHandle *string `json:"sourceHandle,omitempty"`
	TargetHandle *string `json:"targetHandle,omitempty"`
}

type Workflow struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description *string   `json:"description,omitempty"`
	Nodes       []Node    `json:"nodes" gorm:"-"`
	NodesJSON   JSON      `json:"-" gorm:"column:nodes"`
	Edges       []Edge    `json:"edges" gorm:"-"`
	EdgesJSON   JSON      `json:"-" gorm:"column:edges"`
	Active      bool      `json:"active" gorm:"default:false"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName overrides the table name
func (Workflow) TableName() string {
	return "workflows"
}

// BeforeSave hook
func (w *Workflow) BeforeSave(tx *gorm.DB) error {
	// Always set NodesJSON to avoid NULL constraint violations
	if w.Nodes == nil {
		w.NodesJSON = JSON("[]")
	} else {
		nodes, err := json.Marshal(w.Nodes)
		if err != nil {
			return err
		}
		w.NodesJSON = JSON(nodes)
	}

	// Always set EdgesJSON to avoid NULL constraint violations
	if w.Edges == nil {
		w.EdgesJSON = JSON("[]")
	} else {
		edges, err := json.Marshal(w.Edges)
		if err != nil {
			return err
		}
		w.EdgesJSON = JSON(edges)
	}
	return nil
}

// AfterFind hook
func (w *Workflow) AfterFind(tx *gorm.DB) error {
	if len(w.NodesJSON) > 0 && string(w.NodesJSON) != "null" {
		if err := json.Unmarshal(w.NodesJSON, &w.Nodes); err != nil {
			return err
		}
	}
	if len(w.EdgesJSON) > 0 && string(w.EdgesJSON) != "null" {
		if err := json.Unmarshal(w.EdgesJSON, &w.Edges); err != nil {
			return err
		}
	}
	return nil
}

type ExecutionStatus string

const (
	ExecutionStatusRunning ExecutionStatus = "running"
	ExecutionStatusSuccess ExecutionStatus = "success"
	ExecutionStatusError   ExecutionStatus = "error"
	ExecutionStatusWaiting ExecutionStatus = "waiting"
)

type Execution struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	WorkflowID  string                 `json:"workflowId" gorm:"not null;index"`
	Status      ExecutionStatus        `json:"status" gorm:"not null"`
	StartedAt   time.Time              `json:"startedAt" gorm:"not null"`
	FinishedAt  *time.Time             `json:"finishedAt,omitempty"`
	Error       *string                `json:"error,omitempty"`
	Context     map[string]interface{} `json:"context" gorm:"-"`
	ContextJSON JSON                   `json:"-" gorm:"column:context"`
}

// TableName overrides the table name
func (Execution) TableName() string {
	return "executions"
}

// BeforeSave hook
func (e *Execution) BeforeSave(tx *gorm.DB) error {
	// Always set ContextJSON to avoid NULL constraint violations
	if e.Context == nil {
		e.ContextJSON = JSON("{}")
	} else {
		ctx, err := json.Marshal(e.Context)
		if err != nil {
			return err
		}
		e.ContextJSON = JSON(ctx)
	}
	return nil
}

// AfterFind hook
func (e *Execution) AfterFind(tx *gorm.DB) error {
	if len(e.ContextJSON) > 0 && string(e.ContextJSON) != "null" {
		return json.Unmarshal(e.ContextJSON, &e.Context)
	}
	return nil
}

type NodeExecutionLog struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ExecutionID string    `json:"executionId" gorm:"not null;index"`
	NodeID      string    `json:"nodeId" gorm:"not null"`
	Status      string    `json:"status" gorm:"not null"`
	Output      *string   `json:"output,omitempty"`
	Error       *string   `json:"error,omitempty"`
	ExecutedAt  time.Time `json:"executedAt" gorm:"not null"`
}

// TableName overrides the table name
func (NodeExecutionLog) TableName() string {
	return "execution_logs"
}

type ChatSession struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	WorkflowID string    `json:"workflowId" gorm:"not null;index"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName overrides the table name
func (ChatSession) TableName() string {
	return "chat_sessions"
}

type ChatMessage struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"sessionId" gorm:"not null;index"`
	Role      string    `json:"role" gorm:"not null"` // user, assistant, system
	Content   string    `json:"content" gorm:"not null"`
	Timestamp time.Time `json:"timestamp" gorm:"not null"`
}

// TableName overrides the table name
func (ChatMessage) TableName() string {
	return "chat_messages"
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
}

// AI Agent Configuration
type AIAgentConfig struct {
	Model        string  `json:"model"`
	SystemPrompt string  `json:"systemPrompt"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int     `json:"maxTokens"`
}

// HTTP Request Configuration
type HTTPRequestConfig struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"headers,omitempty"`
	Body    map[string]interface{} `json:"body,omitempty"`
}

// Transform Configuration
type TransformConfig struct {
	Expression string `json:"expression"`
}

// Condition Configuration
type ConditionConfig struct {
	Expression string `json:"expression"`
}

// Credential types
type CredentialType string

const (
	CredentialTypeAPIKey     CredentialType = "api_key"
	CredentialTypeOAuth      CredentialType = "oauth"
	CredentialTypeBasicAuth  CredentialType = "basic_auth"
	CredentialTypeBearerToken CredentialType = "bearer_token"
	CredentialTypeCustom     CredentialType = "custom"
)

// Credential stores encrypted user credentials
type Credential struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Type        CredentialType `json:"type" gorm:"not null"`
	Description *string        `json:"description,omitempty"`
	// Value is encrypted at rest
	EncryptedValue string                 `json:"-" gorm:"column:encrypted_value;not null"`
	// DecryptedValue is used temporarily and never stored
	DecryptedValue string                 `json:"value,omitempty" gorm:"-"`
	// Metadata stores additional non-sensitive information
	Metadata     map[string]interface{} `json:"metadata,omitempty" gorm:"-"`
	MetadataJSON JSON                   `json:"-" gorm:"column:metadata"`
	CreatedAt    time.Time              `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time              `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName overrides the table name
func (Credential) TableName() string {
	return "credentials"
}

// BeforeSave hook
func (c *Credential) BeforeSave(tx *gorm.DB) error {
	// Always set MetadataJSON to avoid NULL constraint violations
	if c.Metadata == nil {
		c.MetadataJSON = JSON("{}")
	} else {
		metadata, err := json.Marshal(c.Metadata)
		if err != nil {
			return err
		}
		c.MetadataJSON = JSON(metadata)
	}
	return nil
}

// AfterFind hook
func (c *Credential) AfterFind(tx *gorm.DB) error {
	if len(c.MetadataJSON) > 0 && string(c.MetadataJSON) != "null" {
		if err := json.Unmarshal(c.MetadataJSON, &c.Metadata); err != nil {
			return err
		}
	}
	return nil
}

// WorkflowTemplate represents a pre-built workflow template
type WorkflowTemplate struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description *string   `json:"description,omitempty"`
	Category    string    `json:"category" gorm:"not null"` // e.g., "Customer Support", "Data Processing", "Integration"
	Tags        []string  `json:"tags,omitempty" gorm:"-"`
	TagsJSON    JSON      `json:"-" gorm:"column:tags"`
	Nodes       []Node    `json:"nodes" gorm:"-"`
	NodesJSON   JSON      `json:"-" gorm:"column:nodes"`
	Edges       []Edge    `json:"edges" gorm:"-"`
	EdgesJSON   JSON      `json:"-" gorm:"column:edges"`
	Thumbnail   *string   `json:"thumbnail,omitempty"` // URL or base64 encoded image
	Featured    bool      `json:"featured" gorm:"default:false"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName overrides the table name
func (WorkflowTemplate) TableName() string {
	return "workflow_templates"
}

// BeforeSave hook
func (wt *WorkflowTemplate) BeforeSave(tx *gorm.DB) error {
	// Always set NodesJSON to avoid NULL constraint violations
	if wt.Nodes == nil {
		wt.NodesJSON = JSON("[]")
	} else {
		nodes, err := json.Marshal(wt.Nodes)
		if err != nil {
			return err
		}
		wt.NodesJSON = JSON(nodes)
	}

	// Always set EdgesJSON to avoid NULL constraint violations
	if wt.Edges == nil {
		wt.EdgesJSON = JSON("[]")
	} else {
		edges, err := json.Marshal(wt.Edges)
		if err != nil {
			return err
		}
		wt.EdgesJSON = JSON(edges)
	}

	// Always set TagsJSON to avoid NULL constraint violations
	if wt.Tags == nil {
		wt.TagsJSON = JSON("[]")
	} else {
		tags, err := json.Marshal(wt.Tags)
		if err != nil {
			return err
		}
		wt.TagsJSON = JSON(tags)
	}

	return nil
}

// AfterFind hook
func (wt *WorkflowTemplate) AfterFind(tx *gorm.DB) error {
	if len(wt.NodesJSON) > 0 && string(wt.NodesJSON) != "null" {
		if err := json.Unmarshal(wt.NodesJSON, &wt.Nodes); err != nil {
			return err
		}
	}
	if len(wt.EdgesJSON) > 0 && string(wt.EdgesJSON) != "null" {
		if err := json.Unmarshal(wt.EdgesJSON, &wt.Edges); err != nil {
			return err
		}
	}
	if len(wt.TagsJSON) > 0 && string(wt.TagsJSON) != "null" {
		if err := json.Unmarshal(wt.TagsJSON, &wt.Tags); err != nil {
			return err
		}
	}
	return nil
}
