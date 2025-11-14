package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/workflow-orchestration/backend/database"
	"github.com/workflow-orchestration/backend/models"
	"github.com/workflow-orchestration/backend/utils"
)

type NodeExecutor struct {
	db              *database.DB
	anthropicAPIKey string
}

func NewNodeExecutor(db *database.DB, anthropicAPIKey string) *NodeExecutor {
	return &NodeExecutor{
		db:              db,
		anthropicAPIKey: anthropicAPIKey,
	}
}

// getCredentialValue retrieves and decrypts a credential value by ID
func (ne *NodeExecutor) getCredentialValue(credentialID string) (string, error) {
	if credentialID == "" {
		return "", nil
	}

	credential, err := ne.db.GetCredential(credentialID)
	if err != nil {
		return "", fmt.Errorf("failed to get credential: %w", err)
	}
	if credential == nil {
		return "", fmt.Errorf("credential not found: %s", credentialID)
	}

	// Decrypt the value
	decryptedValue, err := utils.Decrypt(credential.EncryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt credential: %w", err)
	}

	return decryptedValue, nil
}

func (ne *NodeExecutor) Execute(node *models.Node, ctx map[string]interface{}) (interface{}, error) {
	switch node.Type {
	case models.NodeTypeChatTrigger:
		return ne.executeChatTrigger(node, ctx)
	case models.NodeTypeAIAgent:
		return ne.executeAIAgent(node, ctx)
	case models.NodeTypeHTTPRequest:
		return ne.executeHTTPRequest(node, ctx)
	case models.NodeTypeTransform:
		return ne.executeTransform(node, ctx)
	case models.NodeTypeCondition:
		return ne.executeCondition(node, ctx)
	case models.NodeTypeCode:
		return ne.executeCode(node, ctx)
	case models.NodeTypeDelay:
		return ne.executeDelay(node, ctx)
	case models.NodeTypeResponse:
		return ne.executeResponse(node, ctx)
	default:
		return nil, fmt.Errorf("unknown node type: %s", node.Type)
	}
}

func (ne *NodeExecutor) executeChatTrigger(node *models.Node, ctx map[string]interface{}) (interface{}, error) {
	// Chat trigger just passes through the trigger data
	if trigger, ok := ctx["trigger"]; ok {
		return trigger, nil
	}
	return map[string]interface{}{}, nil
}

func (ne *NodeExecutor) executeAIAgent(node *models.Node, ctx map[string]interface{}) (interface{}, error) {
	// Use AI Agent Service instead of calling Anthropic directly
	aiAgentServiceURL := "http://ai-agent:8000"

	// Extract configuration
	provider := "anthropic"
	if p, ok := node.Data["provider"].(string); ok && p != "" {
		provider = p
	}

	model := "claude-3-5-sonnet-20241022"
	if m, ok := node.Data["model"].(string); ok && m != "" {
		model = m
	}

	systemPrompt := ""
	if sp, ok := node.Data["systemPrompt"].(string); ok {
		systemPrompt = ne.interpolateVariables(sp, ctx)
	}

	temperature := 0.7
	if t, ok := node.Data["temperature"].(float64); ok {
		temperature = t
	}

	maxTokens := 4096
	if mt, ok := node.Data["maxTokens"].(float64); ok {
		maxTokens = int(mt)
	}

	// Memory configuration
	memoryEnabled := true
	if me, ok := node.Data["memoryEnabled"].(bool); ok {
		memoryEnabled = me
	}

	maxMemoryMessages := 50
	if mmm, ok := node.Data["maxMemoryMessages"].(float64); ok {
		maxMemoryMessages = int(mmm)
	}

	// MCP Tools configuration
	tools := []map[string]interface{}{}
	if toolsData, ok := node.Data["tools"].([]interface{}); ok {
		for _, t := range toolsData {
			if tool, ok := t.(map[string]interface{}); ok {
				tools = append(tools, tool)
			}
		}
	}

	// Get session ID from context (or generate one from workflow execution)
	sessionID := ""
	if trigger, ok := ctx["trigger"].(map[string]interface{}); ok {
		if sid, ok := trigger["sessionId"].(string); ok {
			sessionID = sid
		}
	}
	if sessionID == "" {
		// Use a session ID based on workflow execution context
		if executionID, ok := ctx["executionId"].(string); ok {
			sessionID = executionID
		}
	}

	// Extract user message from trigger
	userMessage := "Hello"
	if trigger, ok := ctx["trigger"].(map[string]interface{}); ok {
		if message, ok := trigger["message"].(string); ok {
			userMessage = message
		}
	}

	// Step 1: Create or update agent configuration
	agentID := node.ID
	agentConfig := map[string]interface{}{
		"agent_id":            agentID,
		"provider":            provider,
		"model":               model,
		"temperature":         temperature,
		"max_tokens":          maxTokens,
		"system_prompt":       systemPrompt,
		"tools":               tools,
		"memory_enabled":      memoryEnabled,
		"max_memory_messages": maxMemoryMessages,
	}

	createAgentRequest := map[string]interface{}{
		"agent_id": agentID,
		"config":   agentConfig,
	}

	createAgentJSON, err := json.Marshal(createAgentRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent config: %w", err)
	}

	// Create/update agent (idempotent)
	req, err := http.NewRequest("POST", aiAgentServiceURL+"/agents", bytes.NewBuffer(createAgentJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// If agent service is down, return error
		return nil, fmt.Errorf("AI Agent Service unavailable: %w", err)
	}
	resp.Body.Close()

	// Step 2: Chat with agent
	chatRequest := map[string]interface{}{
		"agent_id":   agentID,
		"message":    userMessage,
		"session_id": sessionID,
		"context":    ctx,
	}

	chatJSON, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chat request: %w", err)
	}

	chatReq, err := http.NewRequest("POST", fmt.Sprintf("%s/agents/%s/chat", aiAgentServiceURL, agentID), bytes.NewBuffer(chatJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create chat request: %w", err)
	}
	chatReq.Header.Set("Content-Type", "application/json")

	chatResp, err := client.Do(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to chat with agent: %w", err)
	}
	defer chatResp.Body.Close()

	body, err := io.ReadAll(chatResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if chatResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI Agent Service error (%d): %s", chatResp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract response text
	responseText := ""
	if response, ok := result["response"].(string); ok {
		responseText = response
	}

	// Return in a format compatible with existing workflows
	return map[string]interface{}{
		"text":         responseText,
		"response":     responseText, // alias
		"session_id":   result["session_id"],
		"tool_calls":   result["tool_calls"],
		"metadata":     result["metadata"],
		"fullResponse": result,
	}, nil
}

func (ne *NodeExecutor) executeHTTPRequest(node *models.Node, ctx map[string]interface{}) (interface{}, error) {
	method := "GET"
	if m, ok := node.Data["method"].(string); ok {
		method = m
	}

	url := ""
	if u, ok := node.Data["url"].(string); ok {
		url = ne.interpolateVariables(u, ctx)
	}

	if url == "" {
		return nil, fmt.Errorf("URL is required")
	}

	var reqBody io.Reader
	if method != "GET" && method != "DELETE" {
		if bodyData, ok := node.Data["body"].(map[string]interface{}); ok {
			interpolatedBody := ne.interpolateObject(bodyData, ctx)
			jsonData, err := json.Marshal(interpolatedBody)
			if err != nil {
				return nil, err
			}
			reqBody = bytes.NewBuffer(jsonData)
		}
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	// Add headers
	if headers, ok := node.Data["headers"].(map[string]interface{}); ok {
		for key, val := range headers {
			if strVal, ok := val.(string); ok {
				req.Header.Set(key, ne.interpolateVariables(strVal, ctx))
			}
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON
	var data interface{}
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		json.Unmarshal(body, &data)
	} else {
		data = string(body)
	}

	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return map[string]interface{}{
		"status":     resp.StatusCode,
		"statusText": resp.Status,
		"headers":    headers,
		"data":       data,
	}, nil
}

func (ne *NodeExecutor) executeTransform(node *models.Node, ctx map[string]interface{}) (interface{}, error) {
	// For now, return a simple transformation
	// In a production system, you'd want to use a JS runtime like goja or otto
	expression, ok := node.Data["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("expression is required")
	}

	// This is a simplified version - in production, use a proper JS engine
	result := map[string]interface{}{
		"expression": expression,
		"note":       "Transform executed (JS evaluation not fully implemented in Go version)",
	}

	return result, nil
}

func (ne *NodeExecutor) executeCondition(node *models.Node, ctx map[string]interface{}) (interface{}, error) {
	// Simplified condition evaluation
	expression, ok := node.Data["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("expression is required")
	}

	// Basic condition evaluation (simplified)
	result := false

	// Check for simple includes condition
	if strings.Contains(expression, ".includes(") {
		// Extract the value to check
		re := regexp.MustCompile(`trigger\.message\.includes\(['"](.+?)['"]\)`)
		matches := re.FindStringSubmatch(expression)
		if len(matches) > 1 {
			searchStr := matches[1]
			if trigger, ok := ctx["trigger"].(map[string]interface{}); ok {
				if message, ok := trigger["message"].(string); ok {
					result = strings.Contains(strings.ToLower(message), strings.ToLower(searchStr))
				}
			}
		}
	}

	return map[string]interface{}{
		"condition": result,
		"branch":    map[bool]string{true: "true", false: "false"}[result],
	}, nil
}

func (ne *NodeExecutor) executeCode(node *models.Node, ctx map[string]interface{}) (interface{}, error) {
	code, ok := node.Data["code"].(string)
	if !ok {
		return nil, fmt.Errorf("code is required")
	}

	// In a production system, use a JS runtime like goja or otto
	return map[string]interface{}{
		"code": code,
		"note": "Code execution not fully implemented in Go version",
	}, nil
}

func (ne *NodeExecutor) executeDelay(node *models.Node, ctx map[string]interface{}) (interface{}, error) {
	delayMs := 1000.0
	if d, ok := node.Data["delayMs"].(float64); ok {
		delayMs = d
	}

	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	return map[string]interface{}{
		"delayed": delayMs,
	}, nil
}

func (ne *NodeExecutor) executeResponse(node *models.Node, ctx map[string]interface{}) (interface{}, error) {
	message := ""
	if m, ok := node.Data["message"].(string); ok {
		message = ne.interpolateVariables(m, ctx)
	}

	return map[string]interface{}{
		"message": message,
	}, nil
}

func (ne *NodeExecutor) interpolateVariables(template string, ctx map[string]interface{}) string {
	re := regexp.MustCompile(`\{\{(\w+(?:\.\w+)*)\}\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		path := strings.Trim(match, "{}")
		path = strings.TrimSpace(path)

		value := ne.getNestedValue(ctx, path)
		if value != nil {
			return fmt.Sprintf("%v", value)
		}
		return match
	})
}

func (ne *NodeExecutor) interpolateObject(obj map[string]interface{}, ctx map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range obj {
		switch v := value.(type) {
		case string:
			result[key] = ne.interpolateVariables(v, ctx)
		case map[string]interface{}:
			result[key] = ne.interpolateObject(v, ctx)
		default:
			result[key] = value
		}
	}
	return result
}

func (ne *NodeExecutor) getNestedValue(obj map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(obj)

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}
