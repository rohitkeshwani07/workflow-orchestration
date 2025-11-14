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

	"github.com/workflow-orchestration/backend/models"
)

type NodeExecutor struct {
	anthropicAPIKey string
}

func NewNodeExecutor(anthropicAPIKey string) *NodeExecutor {
	return &NodeExecutor{
		anthropicAPIKey: anthropicAPIKey,
	}
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
	if ne.anthropicAPIKey == "" {
		return nil, fmt.Errorf("Anthropic API key not configured")
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

	// Build messages
	messages := []map[string]interface{}{}

	// Add user input from trigger
	if trigger, ok := ctx["trigger"].(map[string]interface{}); ok {
		if message, ok := trigger["message"].(string); ok {
			messages = append(messages, map[string]interface{}{
				"role":    "user",
				"content": message,
			})
		}
	}

	// If no messages, add a default one
	if len(messages) == 0 {
		messages = append(messages, map[string]interface{}{
			"role":    "user",
			"content": "Hello",
		})
	}

	// Build request
	requestBody := map[string]interface{}{
		"model":       model,
		"max_tokens":  maxTokens,
		"temperature": temperature,
		"messages":    messages,
	}

	if systemPrompt != "" {
		requestBody["system"] = systemPrompt
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", ne.anthropicAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Extract text from response
	text := ""
	if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
		if contentBlock, ok := content[0].(map[string]interface{}); ok {
			if t, ok := contentBlock["text"].(string); ok {
				text = t
			}
		}
	}

	return map[string]interface{}{
		"text":         text,
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
