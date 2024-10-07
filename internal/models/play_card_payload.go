package models

type PlayCardPayload struct {
	PlayerID   string `json:"player_id"`
	CardNumber int    `json:"card_number"`
}