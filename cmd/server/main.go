package main

import (
	"github.com/SimonePesci/The-Mind-Game/internal/handlers"
	"github.com/SimonePesci/The-Mind-Game/internal/utils"
	"github.com/gin-gonic/gin"
)

func main() {

	logger := utils.NewLogger()

	router := gin.Default()

	router.GET("/ws", func(c *gin.Context) {
		handlers.HandleWebSocket(c, logger)
	})

	port := ":8080"
	logger.Infof("Starting server on port %s", port)

	if err := router.Run(port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
