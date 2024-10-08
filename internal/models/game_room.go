package models

import "sync"

// GameRoom represents a game room.
type GameRoom struct {
	ID           string
	Players      map[string]*Player
	Deck         []int
	CurrentRound int
	Lives        int
	Shurikens    int
	Mu           sync.Mutex // Mutex to protect concurrent access
}