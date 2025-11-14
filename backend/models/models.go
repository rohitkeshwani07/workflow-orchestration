package models

import "time"

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

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Node struct {
	ID       string                 `json:"id"`
	Type     NodeType               `json:"type"`
	Position Position               `json:"position"`
	Data     map[string]interface{} `json:"data"`
}

type Edge struct {
	ID           string  `json:"id"`
	Source       string  `json:"source"`
	Target       string  `json:"target"`
	SourceHandle *string `json:"sourceHandle,omitempty"`
	TargetHandle *string `json:"targetHandle,omitempty"`
}

type Workflow struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Nodes       []Node    `json:"nodes"`
	Edges       []Edge    `json:"edges"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ExecutionStatus string

const (
	ExecutionStatusRunning ExecutionStatus = "running"
	ExecutionStatusSuccess ExecutionStatus = "success"
	ExecutionStatusError   ExecutionStatus = "error"
	ExecutionStatusWaiting ExecutionStatus = "waiting"
)

type Execution struct {
	ID         string                 `json:"id"`
	WorkflowID string                 `json:"workflowId"`
	Status     ExecutionStatus        `json:"status"`
	StartedAt  time.Time              `json:"startedAt"`
	FinishedAt *time.Time             `json:"finishedAt,omitempty"`
	Error      *string                `json:"error,omitempty"`
	Context    map[string]interface{} `json:"context"`
}

type NodeExecutionLog struct {
	ID          int       `json:"id"`
	ExecutionID string    `json:"executionId"`
	NodeID      string    `json:"nodeId"`
	Status      string    `json:"status"`
	Output      *string   `json:"output,omitempty"`
	Error       *string   `json:"error,omitempty"`
	ExecutedAt  time.Time `json:"executedAt"`
}

type ChatSession struct {
	ID         string    `json:"id"`
	WorkflowID string    `json:"workflowId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
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
