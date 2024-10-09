package models

// WrongCardPayload contains information about the wrong card played.
type WrongCardPayload struct {
	PlayerID   string `json:"player_id"`
	CardNumber int    `json:"card_number"`
	Position   int    `json:"position"`
	LivesLeft  int    `json:"lives_left"`
}