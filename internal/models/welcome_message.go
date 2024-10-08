package models

// WelcomePayload is the payload for a WELCOME message.
type WelcomePayload struct {
	Message  string `json:"message"`
	PlayerID string `json:"player_id"`
	RoomID   string `json:"room_id"`
}