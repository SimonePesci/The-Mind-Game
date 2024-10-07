package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/SimonePesci/The-Mind-Game/internal/models"
	"github.com/SimonePesci/The-Mind-Game/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	// Allow all origins for simplicity. In production, specify allowed origins.	
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket upgrades the HTTP connection to a WebSocket and handles messages
func HandleWebSocket(c *gin.Context, logger *logrus.Logger) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	playerID := uuid.New().String()

	player := &models.Player{
		ID: playerID,
		Conn: conn,
		Hand: []int{},
	}

	manager := services.GetInstance()
	gameRoom := manager.AddPlayer(player)

	logger.Infof("Player %s connected to room %s", playerID, gameRoom.ID)

	welcomeMsg := models.Message{
		Type: "WELCOME",
		Payload: mustMarshalJSON(map[string]string{
			"message": "Welcome to The Mind!",
			"playerID": playerID,
			"roomID": gameRoom.ID,
		}),
	}

	if err := conn.WriteJSON(welcomeMsg); err != nil {
		logger.Errorf("Failed to send welcome message to player %s: %v", playerID, err)
	}

	for {
		var msg models.Message
		if err := conn.ReadJSON(&msg); err != nil {
			logger.Errorf("Error reading message from player %s: %v", playerID, err)
			break
		}

		logger.Infof("Received message from player %s: %s", playerID, msg.Type)

		switch msg.Type {
		case "PLAY_CARD":
			var payload models.PlayCardPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				logger.Errorf("Invalid PLAY_CARD payload from player %s: %v", playerID, err)
				continue
			}
			manager.HandlePlayCard(gameRoom, payload)

		case "DISCARD_CARD":
			var payload models.DiscardCardPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				logger.Errorf("Invalid DISCARD_CARD payload from player %s: %v", playerID, err)
				continue
			}
			manager.HandleDiscardCard(gameRoom, payload)


			// Handle additional message types here - TODO

		default:
			logger.Warnf("Unknown message type from player %s: %s", playerID, msg.Type)
		}
	}

	// Remove player from the game room upon disconnection
	manager.RemovePlayer(playerID)
	logger.Infof("Player %s disconnected from room %s", playerID, gameRoom.ID)
}

// Helper function to marshal JSON and handle errors
func mustMarshalJSON(v interface{}) json.RawMessage {
	bytes, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	return bytes
}