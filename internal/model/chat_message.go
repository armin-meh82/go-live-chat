package model

import "time"

type ChatMessage struct {
	ID      string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	RoomID  string    `gorm:"not null;index:idx_room_sent_at,priority:1;type:varchar(64)" json:"roomId"`
	UserID  string    `gorm:"not null;type:varchar(64)" json:"userId"`
	Type    string    `gorm:"not null;type:varchar(50)" json:"type"`
	Message string    `gorm:"not null;type:varchar(500)" json:"message"`
	SentAt  time.Time `gorm:"not null;index:idx_room_sent_at,priority:2,sort:desc" json:"sentAt"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
