package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/SimonePesci/The-Mind-Game/internal/models"
	"github.com/SimonePesci/The-Mind-Game/internal/services"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow any origin for simplicity; adjust as needed.
		return true
	},
}

// HandleWebSocket upgrades the HTTP connection to a WebSocket and manages communication.
func HandleWebSocket(w http.ResponseWriter, r *http.Request, logger *logrus.Logger) {
	// Upgrade the HTTP connection to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()
	logger.Info("WebSocket connection established")

	// Create a new player and add them to a game room
	player := models.NewPlayer(conn)
	manager := services.GetGameRoomManager()
	gameRoom := manager.AddPlayer(player)
	logger.Infof("Player %s connected to room %s", player.ID, gameRoom.ID)

	// Send a welcome message to the player
	welcomePayload := models.WelcomePayload{
		Message:  "Welcome to The Mind!",
		PlayerID: player.ID,
		RoomID:   gameRoom.ID,
	}
	if err := sendMessage(conn, "WELCOME", welcomePayload, logger); err != nil {
		return
	}

	// Handle incoming messages from the player
	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			logger.Errorf("Error reading message from player %s: %v", player.ID, err)
			break
		}

		var msg models.Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			logger.Errorf("Invalid message format from player %s: %v", player.ID, err)
			continue
		}

		logger.Infof("Received message from player %s: %s", player.ID, msg.Type)

		// Handle different message types
		switch msg.Type {
		case "PLAY_CARD":
			var payload models.PlayCardPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				logger.Errorf("Invalid PLAY_CARD payload from player %s: %v", player.ID, err)
				continue
			}
			manager.HandlePlayCard(gameRoom, payload, logger)

		case "DISCARD_CARD":
			var payload models.DiscardCardPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				logger.Errorf("Invalid DISCARD_CARD payload from player %s: %v", player.ID, err)
				continue
			}
			manager.HandleDiscardCard(gameRoom, payload, logger)

		default:
			logger.Warnf("Unknown message type from player %s: %s", player.ID, msg.Type)
		}
	}

	// Remove the player from the game room upon disconnection
	manager.RemovePlayer(player.ID)
	logger.Infof("Player %s disconnected from room %s", player.ID, gameRoom.ID)
}

// sendMessage sends a message to the client over the WebSocket connection.
func sendMessage(conn *websocket.Conn, messageType string, payload interface{}, logger *logrus.Logger) error {
	message := models.Message{
		Type:    messageType,
		Payload: nil,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf("Failed to marshal payload: %v", err)
		return err
	}
	message.Payload = payloadBytes

	if err := conn.WriteJSON(message); err != nil {
		logger.Errorf("Failed to send message: %v", err)
		return err
	}
	logger.Infof("Sent %s message", messageType)
	return nil
}