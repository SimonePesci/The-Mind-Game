package models

type DiscardCardPayload struct {
	PlayerID   string `json:"player_id"`
	CardNumber int    `json:"card_number"`
}