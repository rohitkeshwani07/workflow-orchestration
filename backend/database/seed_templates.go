package database

import (
	"github.com/google/uuid"
	"github.com/workflow-orchestration/backend/models"
	"time"
)

// SeedTemplates creates example workflow templates if they don't exist
func (db *DB) SeedTemplates() error {
	// Check if templates already exist
	existing, err := db.GetAllWorkflowTemplates()
	if err != nil {
		return err
	}

	if len(existing) > 0 {
		// Templates already seeded
		return nil
	}

	now := time.Now()
	templates := []models.WorkflowTemplate{
		{
			ID:          uuid.New().String(),
			Name:        "AI Customer Support Bot",
			Description: strPtr("Automated customer support workflow that uses AI to understand and respond to customer inquiries"),
			Category:    "Customer Support",
			Tags:        []string{"ai", "customer-service", "automation"},
			Featured:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
			Nodes: []models.Node{
				{
					ID:   "trigger-1",
					Type: models.NodeTypeChatTrigger,
					Position: models.Position{
						X: 100,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label": "Customer Message",
					},
				},
				{
					ID:   "ai-agent-1",
					Type: models.NodeTypeAIAgent,
					Position: models.Position{
						X: 300,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label":             "AI Support Agent",
						"provider":          "anthropic",
						"model":             "claude-3-5-sonnet-20241022",
						"systemPrompt":      "You are a helpful customer support agent. Be friendly, professional, and resolve customer issues efficiently.",
						"temperature":       0.7,
						"maxTokens":         1000,
						"memoryEnabled":     true,
						"maxMemoryMessages": 10,
					},
				},
				{
					ID:   "response-1",
					Type: models.NodeTypeResponse,
					Position: models.Position{
						X: 500,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label": "Send Response",
					},
				},
			},
			Edges: []models.Edge{
				{
					ID:     "edge-1",
					Source: "trigger-1",
					Target: "ai-agent-1",
				},
				{
					ID:     "edge-2",
					Source: "ai-agent-1",
					Target: "response-1",
				},
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "Data Processing Pipeline",
			Description: strPtr("Extract, transform, and load data from HTTP APIs with conditional processing"),
			Category:    "Data Processing",
			Tags:        []string{"etl", "api", "automation"},
			Featured:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
			Nodes: []models.Node{
				{
					ID:   "trigger-1",
					Type: models.NodeTypeChatTrigger,
					Position: models.Position{
						X: 100,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label": "Manual Trigger",
					},
				},
				{
					ID:   "http-1",
					Type: models.NodeTypeHTTPRequest,
					Position: models.Position{
						X: 300,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label":  "Fetch Data",
						"method": "GET",
						"url":    "https://api.example.com/data",
						"headers": map[string]string{
							"Content-Type": "application/json",
						},
					},
				},
				{
					ID:   "transform-1",
					Type: models.NodeTypeTransform,
					Position: models.Position{
						X: 500,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label":      "Transform Data",
						"expression": "{ \"processed\": data.results.map(r => r.value * 2) }",
					},
				},
				{
					ID:   "condition-1",
					Type: models.NodeTypeCondition,
					Position: models.Position{
						X: 700,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label":      "Check Quality",
						"expression": "data.processed.length > 10",
					},
				},
				{
					ID:   "http-2",
					Type: models.NodeTypeHTTPRequest,
					Position: models.Position{
						X: 900,
						Y: 50,
					},
					Data: map[string]interface{}{
						"label":  "Send to Analytics",
						"method": "POST",
						"url":    "https://api.example.com/analytics",
					},
				},
				{
					ID:   "response-1",
					Type: models.NodeTypeResponse,
					Position: models.Position{
						X: 900,
						Y: 150,
					},
					Data: map[string]interface{}{
						"label": "Complete",
					},
				},
			},
			Edges: []models.Edge{
				{
					ID:     "edge-1",
					Source: "trigger-1",
					Target: "http-1",
				},
				{
					ID:     "edge-2",
					Source: "http-1",
					Target: "transform-1",
				},
				{
					ID:     "edge-3",
					Source: "transform-1",
					Target: "condition-1",
				},
				{
					ID:           "edge-4",
					Source:       "condition-1",
					Target:       "http-2",
					SourceHandle: strPtr("true"),
				},
				{
					ID:           "edge-5",
					Source:       "condition-1",
					Target:       "response-1",
					SourceHandle: strPtr("false"),
				},
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "Content Moderation Bot",
			Description: strPtr("Automatically moderate user-generated content using AI"),
			Category:    "Content Moderation",
			Tags:        []string{"ai", "moderation", "safety"},
			Featured:    false,
			CreatedAt:   now,
			UpdatedAt:   now,
			Nodes: []models.Node{
				{
					ID:   "trigger-1",
					Type: models.NodeTypeChatTrigger,
					Position: models.Position{
						X: 100,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label": "User Content",
					},
				},
				{
					ID:   "ai-agent-1",
					Type: models.NodeTypeAIAgent,
					Position: models.Position{
						X: 300,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label":        "Content Moderator",
						"provider":     "openai",
						"model":        "gpt-4-turbo-preview",
						"systemPrompt": "You are a content moderation AI. Analyze the content and determine if it violates community guidelines. Return a JSON object with 'safe' (boolean) and 'reason' (string).",
						"temperature":  0.3,
						"maxTokens":    500,
					},
				},
				{
					ID:   "condition-1",
					Type: models.NodeTypeCondition,
					Position: models.Position{
						X: 500,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label":      "Is Safe?",
						"expression": "data.safe === true",
					},
				},
				{
					ID:   "http-1",
					Type: models.NodeTypeHTTPRequest,
					Position: models.Position{
						X: 700,
						Y: 50,
					},
					Data: map[string]interface{}{
						"label":  "Approve Content",
						"method": "POST",
						"url":    "https://api.example.com/content/approve",
					},
				},
				{
					ID:   "http-2",
					Type: models.NodeTypeHTTPRequest,
					Position: models.Position{
						X: 700,
						Y: 150,
					},
					Data: map[string]interface{}{
						"label":  "Flag Content",
						"method": "POST",
						"url":    "https://api.example.com/content/flag",
					},
				},
			},
			Edges: []models.Edge{
				{
					ID:     "edge-1",
					Source: "trigger-1",
					Target: "ai-agent-1",
				},
				{
					ID:     "edge-2",
					Source: "ai-agent-1",
					Target: "condition-1",
				},
				{
					ID:           "edge-3",
					Source:       "condition-1",
					Target:       "http-1",
					SourceHandle: strPtr("true"),
				},
				{
					ID:           "edge-4",
					Source:       "condition-1",
					Target:       "http-2",
					SourceHandle: strPtr("false"),
				},
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "Lead Qualification Assistant",
			Description: strPtr("Qualify and route sales leads using AI and conditional logic"),
			Category:    "Sales",
			Tags:        []string{"sales", "crm", "ai", "automation"},
			Featured:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
			Nodes: []models.Node{
				{
					ID:   "trigger-1",
					Type: models.NodeTypeChatTrigger,
					Position: models.Position{
						X: 100,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label": "New Lead",
					},
				},
				{
					ID:   "ai-agent-1",
					Type: models.NodeTypeAIAgent,
					Position: models.Position{
						X: 300,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label":        "Lead Qualifier",
						"provider":     "anthropic",
						"model":        "claude-3-5-sonnet-20241022",
						"systemPrompt": "You are a sales qualification assistant. Ask relevant questions to determine lead quality. Score the lead from 1-10 and provide reasoning.",
						"temperature":  0.7,
						"maxTokens":    800,
						"memoryEnabled": true,
						"maxMemoryMessages": 20,
					},
				},
				{
					ID:   "transform-1",
					Type: models.NodeTypeTransform,
					Position: models.Position{
						X: 500,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label":      "Extract Score",
						"expression": "{ \"score\": parseInt(data.match(/\\d+/)[0]) }",
					},
				},
				{
					ID:   "condition-1",
					Type: models.NodeTypeCondition,
					Position: models.Position{
						X: 700,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label":      "High Quality?",
						"expression": "data.score >= 7",
					},
				},
				{
					ID:   "http-1",
					Type: models.NodeTypeHTTPRequest,
					Position: models.Position{
						X: 900,
						Y: 50,
					},
					Data: map[string]interface{}{
						"label":  "Route to Sales Team",
						"method": "POST",
						"url":    "https://api.example.com/crm/hot-leads",
					},
				},
				{
					ID:   "http-2",
					Type: models.NodeTypeHTTPRequest,
					Position: models.Position{
						X: 900,
						Y: 150,
					},
					Data: map[string]interface{}{
						"label":  "Add to Nurture Campaign",
						"method": "POST",
						"url":    "https://api.example.com/crm/nurture",
					},
				},
			},
			Edges: []models.Edge{
				{
					ID:     "edge-1",
					Source: "trigger-1",
					Target: "ai-agent-1",
				},
				{
					ID:     "edge-2",
					Source: "ai-agent-1",
					Target: "transform-1",
				},
				{
					ID:     "edge-3",
					Source: "transform-1",
					Target: "condition-1",
				},
				{
					ID:           "edge-4",
					Source:       "condition-1",
					Target:       "http-1",
					SourceHandle: strPtr("true"),
				},
				{
					ID:           "edge-5",
					Source:       "condition-1",
					Target:       "http-2",
					SourceHandle: strPtr("false"),
				},
			},
		},
		{
			ID:          uuid.New().String(),
			Name:        "Simple Code Executor",
			Description: strPtr("Execute custom JavaScript code with delays and responses"),
			Category:    "Development",
			Tags:        []string{"code", "developer", "utility"},
			Featured:    false,
			CreatedAt:   now,
			UpdatedAt:   now,
			Nodes: []models.Node{
				{
					ID:   "trigger-1",
					Type: models.NodeTypeChatTrigger,
					Position: models.Position{
						X: 100,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label": "Start",
					},
				},
				{
					ID:   "code-1",
					Type: models.NodeTypeCode,
					Position: models.Position{
						X: 300,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label": "Run Code",
						"code":  "return { result: 'Hello from code node!', timestamp: new Date().toISOString() };",
					},
				},
				{
					ID:   "delay-1",
					Type: models.NodeTypeDelay,
					Position: models.Position{
						X: 500,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label":    "Wait",
						"duration": 2000,
					},
				},
				{
					ID:   "response-1",
					Type: models.NodeTypeResponse,
					Position: models.Position{
						X: 700,
						Y: 100,
					},
					Data: map[string]interface{}{
						"label": "Complete",
					},
				},
			},
			Edges: []models.Edge{
				{
					ID:     "edge-1",
					Source: "trigger-1",
					Target: "code-1",
				},
				{
					ID:     "edge-2",
					Source: "code-1",
					Target: "delay-1",
				},
				{
					ID:     "edge-3",
					Source: "delay-1",
					Target: "response-1",
				},
			},
		},
	}

	// Create all templates
	for _, template := range templates {
		if err := db.CreateWorkflowTemplate(&template); err != nil {
			return err
		}
	}

	return nil
}

func strPtr(s string) *string {
	return &s
}
