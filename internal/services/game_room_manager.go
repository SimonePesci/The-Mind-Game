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
	logger    *logrus.Logger
}

var manager *GameRoomManager
var once sync.Once

// GetGameRoomManager returns the singleton instance of GameRoomManager.
func GetGameRoomManager() *GameRoomManager {
	once.Do(func() {
		manager = &GameRoomManager{
			gameRooms: make(map[string]*models.GameRoom),
			logger:    logrus.New(),
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
		RoundCards:   make([]int, 0),
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
					m.logger.Errorf("Failed to deal card to player %s: %v", player.ID, err)
				} else {
					player.Hand = append(player.Hand, card)
					// Log the dealt card
					m.logger.Infof("Assigned card %d to player %s in room %s", card, player.ID, room.ID)
					
					// Send the dealt card to the player via WebSocket
					m.SendCardToPlayer(player, card, m.logger)
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
			m.logger.Errorf("Failed to deal card to player %s: %v", player.ID, err)
			// Optionally, handle the error
		} else {
			player.Hand = append(player.Hand, card)
			// Log the dealt card
			m.logger.Infof("Assigned card %d to player %s in room %s", card, player.ID, newRoom.ID)
			
			// Send the dealt card to the player via WebSocket
			m.SendCardToPlayer(player, card, m.logger)
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

	// Add the card to the round cards
	room.RoundCards = append(room.RoundCards, payload.CardNumber)

	
	// Broadcast the play action to all players
	m.BroadcastMessage(room, "CARD_PLAYED", payload, logger)
	
	// Validate the sequence of played cards
	position, wrongCard := m.ValidateCardsPlayed(room)
	if position != -1 {
		room.Lives--
		logger.Infof("Player %s played wrong card %d at position %d. Lives left: %d", payload.PlayerID, wrongCard, position, room.Lives)

		// Notify all players about the wrong card
		wrongCardPayload := models.WrongCardPayload{
			PlayerID:   payload.PlayerID,
			CardNumber: wrongCard,
			Position:   position,
			LivesLeft:  room.Lives,
		}
		m.BroadcastMessage(room, "WRONG_CARD", wrongCardPayload, logger)

		if room.Lives < 0 {
			logger.Infof("Game Over in room %s. All lives lost.", room.ID)
			// Handle game over logic here
			return
		}

	}

	room.Mu.Unlock()

	// Check if all players have played their cards
	if m.AllPlayersHavePlayed(room) {

		// TODO: Implement logic to determine if the cards played are valid

		// Start the next round
		if err := m.StartNextRound(room, logger); err != nil {
			logger.Errorf("Failed to start next round in room %s: %v", room.ID, err)
		}
	}
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

	// Check if all players have taken their discard actions
	if m.AllPlayersHaveDiscarded(room) {
		// Start the next round
		if err := m.StartNextRound(room, logger); err != nil {
			logger.Errorf("Failed to start next round in room %s: %v", room.ID, err)
		}
	}
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

// DealCardsForRound deals cards to all players in the room based on the current round.
func (m *GameRoomManager) DealCardsForRound(room *models.GameRoom, logger *logrus.Logger) error {
	logger.Infof("Entering DealCardsForRound for room %s", room.ID)

    room.Mu.Lock()
    defer room.Mu.Unlock()

    numberOfCards := room.CurrentRound
    logger.Infof("Dealing %d cards to each player in room %s", numberOfCards, room.ID)

    // Log the number of players
    numPlayers := len(room.Players)
    logger.Infof("Number of players in room %s: %d", room.ID, numPlayers)

    if numPlayers == 0 {
        logger.Warnf("No players found in room %s. Skipping card dealing.", room.ID)
        return fmt.Errorf("no players in room %s", room.ID)
    }

    for _, player := range room.Players {
        logger.Infof("Dealing cards to player %s in room %s", player.ID, room.ID)
        var dealtCards []int
        for i := 0; i < numberOfCards; i++ {
            card, err := m.dealCard(room)
            if err != nil {
                logger.Errorf("Failed to deal card to player %s: %v", player.ID, err)
                return err
            }
            player.Hand = append(player.Hand, card)
            dealtCards = append(dealtCards, card)
            logger.Infof("Dealt card %d to player %s in room %s", card, player.ID, room.ID)
        }

        // Send the dealt cards to the player
        m.SendCardsToPlayer(player, dealtCards, logger)
        logger.Infof("Sent %d new cards to player %s in room %s", numberOfCards, player.ID, room.ID)
    }

    logger.Infof("Completed DealCardsForRound for room %s", room.ID)
    return nil
}

// SendCardsToPlayer sends multiple cards to the player.
func (m *GameRoomManager) SendCardsToPlayer(player *models.Player, cards []int, logger *logrus.Logger) {
	cardPayload := models.NewCardsPayload{
		CardNumbers: cards,
	}

	message := models.Message{
		Type:    "NEW_CARDS",
		Payload: nil,
	}

	payloadBytes, err := json.Marshal(cardPayload)
	if err != nil {
		logger.Errorf("Failed to marshal NEW_CARDS payload for player %s: %v", player.ID, err)
		return
	}
	message.Payload = payloadBytes

	if err := player.Conn.WriteJSON(message); err != nil {
		logger.Errorf("Failed to send NEW_CARDS message to player %s: %v", player.ID, err)
	}
}

// StartNextRound transitions the game to the next round and deals new cards.
func (m *GameRoomManager) StartNextRound(room *models.GameRoom, logger *logrus.Logger) error {
	logger.Infof("Entering StartNextRound for room %s", room.ID)

    room.CurrentRound++
    logger.Infof("Incremented to round %d in room %s", room.CurrentRound, room.ID)

    // Reset RoundCards for the new round
    room.RoundCards = make([]int, 0)
    logger.Infof("RoundCards reset for room %s", room.ID)

    // Call DealCardsForRound and log the result
    err := m.DealCardsForRound(room, logger)
    if err != nil {
        logger.Errorf("DealCardsForRound failed for room %s: %v", room.ID, err)
        return err
    }

    logger.Infof("Successfully started round %d in room %s", room.CurrentRound, room.ID)
    return nil
}

// AllPlayersHavePlayed checks if all players have played their cards for the current round.
func (m *GameRoomManager) AllPlayersHavePlayed(room *models.GameRoom) bool {
	// Implement logic to determine if all players have played.
	// This can be tracked using additional fields in GameRoom.
	// For simplicity, returning true.

	for _, player := range room.Players {
		if len(player.Hand) > 0 {
			return false
		}
	}

	return true
}

// AllPlayersHaveDiscarded checks if all players have discarded their cards for the current round.
func (m *GameRoomManager) AllPlayersHaveDiscarded(room *models.GameRoom) bool {
	// Implement logic to determine if all players have discarded.
	// This can be tracked using additional fields in GameRoom.
	// For simplicity, returning true.
	return true
}

// ValidateCardsPlayed validates the cards played in the current round.
// It returns the position and value of the first wrong card if invalid, otherwise returns -1 and 0.
func (m *GameRoomManager) ValidateCardsPlayed(room *models.GameRoom) (int, int) {
	currentCard := 0
	for idx, card := range room.RoundCards {
		if card > currentCard {
			currentCard = card
		} else {
			return idx, card
		}
	}
	return -1, 0
}