package service

import (
	"testing"

	"github.com/armin-meh82/live-chat/internal/chat"
	"github.com/armin-meh82/live-chat/internal/dto"
)

func TestProcessMessageValidation(t *testing.T) {
	hub := chat.NewHub()
	client := chat.NewClient(nil, "room-1", "test-user")
	hub.Register(client)
	defer hub.Unregister(client)

	chatService := NewChatService(
		hub,
		&fakeMessageRepository{},
	)

	tests := []struct {
		name    string
		message dto.IncomingMessageDTO
		wantErr bool
	}{
		{
			name: "valid message",
			message: dto.IncomingMessageDTO{
				Type:    "chat.message",
				Message: "Hello",
			},
		},
		{
			name: "empty message",
			message: dto.IncomingMessageDTO{
				Type:    "chat.message",
				Message: "   ",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			message: dto.IncomingMessageDTO{
				Type:    "unknown",
				Message: "Hello",
			},
			wantErr: true,
		},
		{
			name: "message too long",
			message: dto.IncomingMessageDTO{
				Type:    "chat.message",
				Message: string(make([]byte, 501)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := chatService.ProcessMessage(
				"room-1",
				"test-user",
				tt.message,
			)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
