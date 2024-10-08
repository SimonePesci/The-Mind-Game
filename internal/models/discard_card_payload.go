package models

// DiscardCardPayload is the payload for a DISCARD_CARD message.
type DiscardCardPayload struct {
	PlayerID   string `json:"player_id"`
	CardNumber int    `json:"card_number"`
}