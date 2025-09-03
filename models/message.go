package models

type Message struct {
	ClientID string `json:"client_id,omitempty"`
	Text     string `json:"message,omitempty"`
}
