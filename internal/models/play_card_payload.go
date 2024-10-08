package models

// PlayCardPayload is the payload for a PLAY_CARD message.
type PlayCardPayload struct {
	PlayerID   string `json:"player_id"`
	CardNumber int    `json:"card_number"`
}