package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/workflow-orchestration/backend/database"
	"github.com/workflow-orchestration/backend/engine"
	"github.com/workflow-orchestration/backend/models"
)

type ChatSession struct {
	ID         string
	WorkflowID string
	Conn       *websocket.Conn
}

type ChatWebSocketServer struct {
	db       *database.DB
	engine   *engine.WorkflowEngine
	sessions map[string]*ChatSession
	mu       sync.RWMutex
}

func NewChatWebSocketServer(db *database.DB, engine *engine.WorkflowEngine) *ChatWebSocketServer {
	return &ChatWebSocketServer{
		db:       db,
		engine:   engine,
		sessions: make(map[string]*ChatSession),
	}
}

func (cws *ChatWebSocketServer) HandleConnection(conn *websocket.Conn) {
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			// Clean up session
			cws.cleanupConnection(conn)
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}

		if err := cws.handleMessage(conn, msg); err != nil {
			log.Printf("Handle message error: %v", err)
			cws.sendError(conn, err.Error())
		}
	}
}

func (cws *ChatWebSocketServer) handleMessage(conn *websocket.Conn, msg map[string]interface{}) error {
	msgType, ok := msg["type"].(string)
	if !ok {
		return nil
	}

	switch msgType {
	case "start_session":
		return cws.startSession(conn, msg)
	case "chat_message":
		return cws.handleChatMessage(conn, msg)
	}

	return nil
}

func (cws *ChatWebSocketServer) startSession(conn *websocket.Conn, msg map[string]interface{}) error {
	workflowID, ok := msg["workflowId"].(string)
	if !ok {
		return cws.sendError(conn, "workflowId is required")
	}

	// Verify workflow exists
	workflow, err := cws.db.GetWorkflow(workflowID)
	if err != nil {
		return err
	}
	if workflow == nil {
		return cws.sendError(conn, "Workflow not found")
	}

	sessionID := uuid.New().String()
	if err := cws.db.CreateChatSession(sessionID, workflowID); err != nil {
		return err
	}

	session := &ChatSession{
		ID:         sessionID,
		WorkflowID: workflowID,
		Conn:       conn,
	}

	cws.mu.Lock()
	cws.sessions[sessionID] = session
	cws.mu.Unlock()

	return cws.send(conn, map[string]interface{}{
		"type":       "session_started",
		"sessionId":  sessionID,
		"workflowId": workflowID,
	})
}

func (cws *ChatWebSocketServer) handleChatMessage(conn *websocket.Conn, msg map[string]interface{}) error {
	sessionID, ok := msg["sessionId"].(string)
	if !ok {
		return cws.sendError(conn, "sessionId is required")
	}

	content, ok := msg["content"].(string)
	if !ok {
		return cws.sendError(conn, "content is required")
	}

	cws.mu.RLock()
	session, exists := cws.sessions[sessionID]
	cws.mu.RUnlock()

	if !exists {
		return cws.sendError(conn, "Session not found")
	}

	// Save user message
	userMessage := &models.ChatMessage{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	}
	if err := cws.db.AddChatMessage(userMessage); err != nil {
		return err
	}

	// Send acknowledgment
	if err := cws.send(conn, map[string]interface{}{
		"type":    "message_received",
		"message": userMessage,
	}); err != nil {
		return err
	}

	// Execute workflow
	execution, err := cws.engine.ExecuteFromTrigger(session.WorkflowID, map[string]interface{}{
		"message":   content,
		"sessionId": sessionID,
	})

	if err != nil {
		log.Printf("Workflow execution error: %v", err)
		errorMessage := &models.ChatMessage{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Role:      "assistant",
			Content:   "Error: " + err.Error(),
			Timestamp: time.Now(),
		}
		cws.db.AddChatMessage(errorMessage)

		return cws.send(conn, map[string]interface{}{
			"type":    "chat_response",
			"message": errorMessage,
			"error":   err.Error(),
		})
	}

	// Extract response from execution context
	responseContent := "Workflow executed successfully"

	// Check for response node output
	if responseData, ok := execution.Context["response"].(map[string]interface{}); ok {
		if msg, ok := responseData["message"].(string); ok {
			responseContent = msg
		}
	} else {
		// Look for AI agent response
		for key, value := range execution.Context {
			if key != "trigger" {
				if data, ok := value.(map[string]interface{}); ok {
					if text, ok := data["text"].(string); ok && text != "" {
						responseContent = text
						break
					}
					if msg, ok := data["message"].(string); ok && msg != "" {
						responseContent = msg
						break
					}
				}
			}
		}
	}

	// Save assistant message
	assistantMessage := &models.ChatMessage{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   responseContent,
		Timestamp: time.Now(),
	}
	if err := cws.db.AddChatMessage(assistantMessage); err != nil {
		return err
	}

	// Send response
	return cws.send(conn, map[string]interface{}{
		"type":        "chat_response",
		"message":     assistantMessage,
		"executionId": execution.ID,
	})
}

func (cws *ChatWebSocketServer) send(conn *websocket.Conn, data interface{}) error {
	return conn.WriteJSON(data)
}

func (cws *ChatWebSocketServer) sendError(conn *websocket.Conn, errorMsg string) error {
	return cws.send(conn, map[string]interface{}{
		"type":  "error",
		"error": errorMsg,
	})
}

func (cws *ChatWebSocketServer) cleanupConnection(conn *websocket.Conn) {
	cws.mu.Lock()
	defer cws.mu.Unlock()

	for sessionID, session := range cws.sessions {
		if session.Conn == conn {
			delete(cws.sessions, sessionID)
			log.Printf("Session %s closed", sessionID)
			break
		}
	}
}

func (cws *ChatWebSocketServer) GetActiveSessions() int {
	cws.mu.RLock()
	defer cws.mu.RUnlock()
	return len(cws.sessions)
}
