package models

import "github.com/gorilla/websocket"

type Player struct {
	ID   string
	Conn *websocket.Conn
	Hand []int
	RoomID string
}