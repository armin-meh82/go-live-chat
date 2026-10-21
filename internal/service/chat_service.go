package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/armin-meh82/live-chat/internal/chat"
	"github.com/armin-meh82/live-chat/internal/dto"
	"github.com/armin-meh82/live-chat/internal/repository"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrInvalidRoomID  = errors.New("invalid room ID")
)

type ChatService struct {
	hub  *chat.Hub
	repo repository.MessageRepository
}

func NewChatService(
	hub *chat.Hub,
	repo repository.MessageRepository,
) *ChatService {
	return &ChatService{
		hub:  hub,
		repo: repo,
	}
}

func (s *ChatService) ProcessMessage(
	roomID string,
	userID string,
	incoming dto.IncomingMessageDTO,
) error {
	if !ValidRoomID(roomID) {
		return ErrInvalidRoomID
	}

	if incoming.Type != "chat.message" ||
		strings.TrimSpace(incoming.Message) == "" ||
		utf8.RuneCountInString(incoming.Message) > 500 {
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

	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	if err := s.repo.Save(ctx, message); err != nil {
		return err
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
