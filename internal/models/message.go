package models

import "encoding/json"

// Message represents a message exchanged between client and server.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}