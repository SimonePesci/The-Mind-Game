package models

// NewCardPayload is the payload for a NEW_CARD message.
type NewCardPayload struct {
	CardNumber int `json:"card_number"`
}