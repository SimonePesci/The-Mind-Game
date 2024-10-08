package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/SimonePesci/The-Mind-Game/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// GameRoomManager manages all game rooms.
type GameRoomManager struct {
	gameRooms map[string]*models.GameRoom
	mu        sync.Mutex
}

var manager *GameRoomManager
var once sync.Once

// GetGameRoomManager returns the singleton instance of GameRoomManager.
func GetGameRoomManager() *GameRoomManager {
	once.Do(func() {
		manager = &GameRoomManager{
			gameRooms: make(map[string]*models.GameRoom),
		}
	})
	return manager
}

// CreateGameRoom initializes a new game room and adds it to the manager.
func (m *GameRoomManager) CreateGameRoom() *models.GameRoom {
	roomID := uuid.New().String()
	deck := initializeDeck()

	gameRoom := &models.GameRoom{
		ID:           roomID,
		Players:      make(map[string]*models.Player),
		Deck:         deck,
		CurrentRound: 1,
		Lives:        3,
		Shurikens:    1,
	}

	m.gameRooms[roomID] = gameRoom
	return gameRoom
}

// AddPlayer adds a player to an existing game room or creates a new one if none are available.
func (m *GameRoomManager) AddPlayer(player *models.Player) *models.GameRoom {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Assign to the first available room that isn't full
	for _, room := range m.gameRooms {
		room.Mu.Lock()
		if len(room.Players) < 4 {
			room.Players[player.ID] = player
			player.RoomID = room.ID

			if room.CurrentRound == 1 {
				card, err := m.dealCard(room)
				if err != nil {
					// logger.Errorf("Failed to deal card to player %s: %v", player.ID, err)
				} else {
					player.Hand = append(player.Hand, card)
					// logger.Infof("Assigned card %d to player %s in room %s", card, player.ID, room.ID)
				}
			}

			room.Mu.Unlock()
			return room
		}
		room.Mu.Unlock()
	}

	// If no available room, create a new one and assign a card to the player
	newRoom := m.CreateGameRoom()
	newRoom.Mu.Lock()
	newRoom.Players[player.ID] = player
	player.RoomID = newRoom.ID

	// Assign a card if currentRound is 1
	if newRoom.CurrentRound == 1 {
		card, err := m.dealCard(newRoom)
		if err != nil {
			// logger.Errorf("Failed to deal card to player %s: %v", player.ID, err)
			// Optionally, handle the error
		} else {
			player.Hand = append(player.Hand, card)
			// logger.Infof("Assigned card %d to player %s in room %s", card, player.ID, newRoom.ID)
			// Optionally, send a message to the player about the new card
			// Example: m.SendCardToPlayer(player, card, logger)
		}
	}


	newRoom.Mu.Unlock()
	return newRoom
}

// RemovePlayer removes a player from the game room.
func (m *GameRoomManager) RemovePlayer(playerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, room := range m.gameRooms {
		room.Mu.Lock()
		if _, exists := room.Players[playerID]; exists {
			delete(room.Players, playerID)
			if len(room.Players) == 0 {
				// Remove the room if it's empty
				delete(m.gameRooms, room.ID)
			}
			room.Mu.Unlock()
			return
		}
		room.Mu.Unlock()
	}
}

// initializeDeck creates and shuffles a deck of cards numbered 1 to 100.
func initializeDeck() []int {
	deck := make([]int, 100)
	for i := 0; i < 100; i++ {
		deck[i] = i + 1
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	return deck
}

// dealCard draws a card from the room's deck and returns it.
// Returns an error if the deck is empty.
func (m *GameRoomManager) dealCard(room *models.GameRoom) (int, error) {
	if len(room.Deck) == 0 {
		return 0, fmt.Errorf("deck is empty in room %s", room.ID)
	}
	// Draw the top card (last element for efficiency)
	card := room.Deck[len(room.Deck)-1]
	room.Deck = room.Deck[:len(room.Deck)-1]
	return card, nil
}

// SendCardToPlayer sends a message to the player with the newly assigned card.
func (m *GameRoomManager) SendCardToPlayer(player *models.Player, card int, logger *logrus.Logger) {
	cardPayload := models.NewCardPayload{
		CardNumber: card,
	}

	message := models.Message{
		Type:    "NEW_CARD",
		Payload: nil,
	}

	payloadBytes, err := json.Marshal(cardPayload)
	if err != nil {
		logger.Errorf("Failed to marshal NEW_CARD payload for player %s: %v", player.ID, err)
		return
	}
	message.Payload = payloadBytes

	if err := player.Conn.WriteJSON(message); err != nil {
		logger.Errorf("Failed to send NEW_CARD message to player %s: %v", player.ID, err)
	}
}


// HandlePlayCard processes a player's card play action.
func (m *GameRoomManager) HandlePlayCard(room *models.GameRoom, payload models.PlayCardPayload, logger *logrus.Logger) {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	player, exists := room.Players[payload.PlayerID]
	if !exists {
		logger.Errorf("Player %s not found in room %s", payload.PlayerID, room.ID)
		return
	}

	// Validate that the player has the card
	hasCard := false
	for i, card := range player.Hand {
		if card == payload.CardNumber {
			// Remove the card from player's hand
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			hasCard = true
			break
		}
	}
	if !hasCard {
		logger.Errorf("Player %s does not have card %d", payload.PlayerID, payload.CardNumber)
		return
	}

	// TODO: Implement game logic to check if the card played is valid

	// Broadcast the play action to all players
	m.BroadcastMessage(room, "CARD_PLAYED", payload, logger)
}

// HandleDiscardCard processes a player's card discard action.
func (m *GameRoomManager) HandleDiscardCard(room *models.GameRoom, payload models.DiscardCardPayload, logger *logrus.Logger) {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	player, exists := room.Players[payload.PlayerID]
	if !exists {
		logger.Errorf("Player %s not found in room %s", payload.PlayerID, room.ID)
		return
	}

	if room.Shurikens <= 0 {
		logger.Errorf("No shurikens left to discard in room %s", room.ID)
		return
	}

	// Validate that the player has the card
	hasCard := false
	for i, card := range player.Hand {
		if card == payload.CardNumber {
			// Remove the card from player's hand
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			hasCard = true
			break
		}
	}
	if !hasCard {
		logger.Errorf("Player %s does not have card %d", payload.PlayerID, payload.CardNumber)
		return
	}

	// Decrease shurikens
	room.Shurikens--

	// Broadcast the discard action to all players
	m.BroadcastMessage(room, "CARD_DISCARDED", payload, logger)
}



// BroadcastMessage sends a message to all players in a room.
func (m *GameRoomManager) BroadcastMessage(room *models.GameRoom, messageType string, payload interface{}, logger *logrus.Logger) {
	message := models.Message{
		Type:    messageType,
		Payload: nil,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf("Failed to marshal payload: %v", err)
		return
	}
	message.Payload = payloadBytes

	for _, player := range room.Players {
		if err := player.Conn.WriteJSON(message); err != nil {
			logger.Errorf("Failed to send message to player %s: %v", player.ID, err)
			// Optionally handle disconnection
		}
	}
}