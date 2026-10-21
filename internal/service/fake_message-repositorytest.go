package service

import (
	"context"

	"github.com/armin-meh82/live-chat/internal/dto"
)

type fakeMessageRepository struct{}

func (f *fakeMessageRepository) Save(
    ctx context.Context,
    message dto.ChatMessageDTO,
) error {
    return nil
}