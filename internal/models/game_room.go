package models

import "sync"

type GameRoom struct {
	ID           string
	Players      map[string]*Player
	Deck         []int
	CurrentRound int
	Lives        int
	Shurikens    int

	Mu sync.Mutex // Mutex to protect concurrent access

}