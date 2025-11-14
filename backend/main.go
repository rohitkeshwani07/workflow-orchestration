package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/workflow-orchestration/backend/database"
	"github.com/workflow-orchestration/backend/engine"
	"github.com/workflow-orchestration/backend/handlers"
	ws "github.com/workflow-orchestration/backend/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

func main() {
	// Load environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")

	// Initialize database
	db, err := database.NewFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize workflow engine
	workflowEngine := engine.NewWorkflowEngine(db, anthropicAPIKey)

	// Initialize handlers
	handler := handlers.NewHandler(db, workflowEngine)

	// Initialize WebSocket server
	chatWS := ws.NewChatWebSocketServer(db, workflowEngine)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		workflows := api.Group("/workflows")
		{
			workflows.GET("", handler.GetAllWorkflows)
			workflows.GET("/:id", handler.GetWorkflow)
			workflows.POST("", handler.CreateWorkflow)
			workflows.PUT("/:id", handler.UpdateWorkflow)
			workflows.DELETE("/:id", handler.DeleteWorkflow)
			workflows.POST("/:id/execute", handler.ExecuteWorkflow)
			workflows.GET("/:id/executions", handler.GetWorkflowExecutions)
			workflows.GET("/:id/executions/:executionId", handler.GetExecutionDetails)
		}
	}

	// WebSocket route
	router.GET("/ws/chat", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		go chatWS.HandleConnection(conn)
	})

	// Start server
	log.Printf("🚀 Server running on http://localhost:%s", port)
	log.Printf("📡 WebSocket available at ws://localhost:%s/ws/chat", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
