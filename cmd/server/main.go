package main

import (
	"github.com/SimonePesci/The-Mind-Game/internal/handlers"
	"github.com/SimonePesci/The-Mind-Game/internal/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Logger
	logger := utils.NewLogger()
	logger.Info("Logger initialized")

	// Initialize Gin router
	router := gin.Default()

	// Set up the WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		handlers.HandleWebSocket(c.Writer, c.Request, logger)
	})

	port := ":8080"
	logger.Infof("Starting server on port %s", port)

	if err := router.Run(port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
