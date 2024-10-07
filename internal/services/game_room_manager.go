package services

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/SimonePesci/The-Mind-Game/internal/models"
	"github.com/google/uuid"
)

type GameRoomManager struct {
	gameRooms map[string]*models.GameRoom
	mu sync.Mutex
}

var manager = &GameRoomManager{
	gameRooms: make(map[string]*models.GameRoom),
}

func GetInstance() *GameRoomManager{
	return manager
}

// CreateGameRoom initializes a new game room and adds it to the manager
func (m *GameRoomManager) CreateGameRoom() *models.GameRoom {
	m.mu.Lock()
	defer m.mu.Unlock()

	roomID := uuid.New().String()
	deck := initializeDeck()

	gameRoom := &models.GameRoom{
		ID: roomID,
		Players: make(map[string]*models.Player),
		Deck: deck,
		CurrentRound: 1,
		Lives: 3,
		Shurikens: 1,
	}

	m.gameRooms[roomID] = gameRoom
	return gameRoom

}

// AddPlayer adds a player to an existing game room or creates a new one if none are available
func (m *GameRoomManager) AddPlayer(player *models.Player) *models.GameRoom {
	m.mu.Lock()
	defer m.mu.Unlock()

	// For simplicity, assign to the first available room that isn't full
	for _, room := range m.gameRooms {
		if len(room.Players) < 4 {
			room.Mu.Lock()
			room.Players[player.ID] = player
			player.RoomID = room.ID
			room.Mu.Unlock()
			return room
		}
	}

	// If no available room, create a new one
	newRoom := m.CreateGameRoom()
	newRoom.Mu.Lock()
	newRoom.Players[player.ID] = player
	player.RoomID = newRoom.ID
	newRoom.Mu.Unlock()
	return newRoom
}

// Remove a player from a Room
func (m *GameRoomManager) RemovePlayer(playerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, room := range m.gameRooms {
		room.Mu.Lock()
		if _, exists := room.Players[playerID]; exists {
			delete(room.Players, playerID)
			room.Mu.Unlock()
			// if the room is empty, we can delete the room
			if len(room.Players) == 0 {
				delete(m.gameRooms, room.ID)
			}
			return
		}
		room.Mu.Unlock()
	}
}

func initializeDeck() []int {
	deck := make([]int, 100)
	for i := 0; i < 100; i++ {
		deck[i] = i + 1
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deck), func(i,j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	return deck
}

// HandlePlayCard processes a player's card play action
func (m *GameRoomManager) HandlePlayCard(room *models.GameRoom, payload models.PlayCardPayload) {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	player, exists := room.Players[payload.PlayerID]
	if !exists {
		log.Printf("Player %s not found in room %s", payload.PlayerID, room.ID)
		return
	}

	// Validation, necessary?
	
	// Remove the card from player's hand
	for i, card := range player.Hand {
		if card == payload.CardNumber {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			break
		}
	}

	// Broadcast play action
	m.BroadcastMessage(room, "CARD_PLAYED", payload)
}

// HandleDiscardCard processes a player's card discard action
func (m *GameRoomManager) HandleDiscardCard(room *models.GameRoom, payload models.DiscardCardPayload) {
	room.Mu.Lock()
	defer room.Mu.Unlock()

	player, exists := room.Players[payload.PlayerID]
	if !exists {
		log.Printf("Player %s not found in room %s" , payload.PlayerID, room.ID)
		return
	}

	if room.Shurikens >= 0 {
		log.Printf("No Shurikens left to discard")
		return
	}

	// if no issues to handle, discard the card
	for i, card := range player.Hand {
		if card == payload.CardNumber {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]... )
			break
		}
	}

	room.Shurikens--

	m.BroadcastMessage(room, "CARD_DISCARDED", payload)
}

// Broadcast a message to all players in a room
func (m *GameRoomManager) BroadcastMessage(room *models.GameRoom, messageType string, payload interface{}) {
	message := models.Message{
		Type: messageType,
		Payload: nil,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return
	}

	message.Payload = payloadBytes

	for _, player := range room.Players {
		err := player.Conn.WriteJSON(message)
		if err != nil {
			log.Printf("Failed to send message to player %s: %v", player.ID, err)
			// handle disconnection?
		}
	}
}