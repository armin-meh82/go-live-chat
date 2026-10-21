package dto

import "time"

type ChatMessageDTO struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"roomId"`
	UserID    string    `json:"userId"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	SentAt    time.Time `json:"sentAt"`
}