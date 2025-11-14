package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/workflow-orchestration/credentials-service/crypto"
	"github.com/workflow-orchestration/credentials-service/database"
	"github.com/workflow-orchestration/credentials-service/handlers"
)

func main() {
	// Load environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	// Initialize encryption
	if err := crypto.InitEncryption(); err != nil {
		log.Fatalf("Failed to initialize encryption: %v", err)
	}

	// Initialize database
	db, err := database.NewFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize handlers
	handler := handlers.NewHandler(db)

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
			"status": "ok",
			"service": "credentials",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		credentials := api.Group("/credentials")
		{
			credentials.GET("", handler.GetAllCredentials)
			credentials.GET("/:id", handler.GetCredential)
			credentials.POST("", handler.CreateCredential)
			credentials.PUT("/:id", handler.UpdateCredential)
			credentials.DELETE("/:id", handler.DeleteCredential)
		}
	}

	// Start server
	log.Printf("🔐 Credentials Service running on http://localhost:%s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
