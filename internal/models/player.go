package models

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Player represents a player in the game.
type Player struct {
	ID     string
	Conn   *websocket.Conn
	Hand   []int
	RoomID string
}

// NewPlayer creates a new player with a unique ID.
func NewPlayer(conn *websocket.Conn) *Player {
	return &Player{
		ID:   uuid.New().String(),
		Conn: conn,
		Hand: []int{},
	}
}