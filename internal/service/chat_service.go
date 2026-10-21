package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/armin-meh82/live-chat/internal/chat"
	"github.com/armin-meh82/live-chat/internal/dto"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrInvalidRoomID  = errors.New("invalid room ID")
)

type ChatService struct {
	hub *chat.Hub
}

func NewChatService(hub *chat.Hub) *ChatService {
	return &ChatService{hub: hub}
}

func (s *ChatService) ProcessMessage(
	roomID string,
	userID string,
	incoming dto.IncomingMessageDTO,
) error {
	if !ValidRoomID(roomID) {
		return ErrInvalidRoomID
	}

	if incoming.Type != "chat.message" {
		return ErrInvalidMessage
	}

	if strings.TrimSpace(incoming.Message) == "" {
		return ErrInvalidMessage
	}

	if utf8.RuneCountInString(incoming.Message) > 500 {
		return ErrInvalidMessage
	}

	id, err := newID()
	if err != nil {
		return errors.New("could not create message")
	}

	message := dto.ChatMessageDTO{
		ID:      id,
		RoomID:  roomID,
		UserID:  userID,
		Type:    incoming.Type,
		Message: incoming.Message,
		SentAt:  time.Now().UTC(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		return errors.New("could not encode message")
	}

	s.hub.Broadcast(roomID, data)
	return nil
}

func ValidRoomID(roomID string) bool {
	if len(roomID) == 0 || len(roomID) > 64 {
		return false
	}

	for _, char := range roomID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}

func newID() (string, error) {
	data := make([]byte, 16)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return hex.EncodeToString(data), nil
}
