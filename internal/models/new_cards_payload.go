package models

// NewCardsPayload is the payload for a NEW_CARDS message.
type NewCardsPayload struct {
	CardNumbers []int `json:"card_numbers"`
}