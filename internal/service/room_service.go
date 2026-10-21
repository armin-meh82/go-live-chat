package service

import (
	"github.com/armin-meh82/live-chat/internal/chat"
	"github.com/armin-meh82/live-chat/internal/dto"
)

type RoomService struct {
	hub *chat.Hub
}

func NewRoomService(hub *chat.Hub) *RoomService {
	return &RoomService{hub: hub}
}

func (s *RoomService) GetRoom(
	roomID string,
) (dto.RoomResponseDTO, error) {
	if !ValidRoomID(roomID) {
		return dto.RoomResponseDTO{}, ErrInvalidRoomID
	}

	count, exists := s.hub.RoomInfo(roomID)

	return dto.RoomResponseDTO{
		RoomID:         roomID,
		ConnectedUsers: count,
		Active:         exists,
	}, nil
}
