package repository

import (
	"context"
	"fmt"

	"github.com/armin-meh82/live-chat/internal/dto"
	"github.com/armin-meh82/live-chat/internal/model"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Save(ctx context.Context, message dto.ChatMessageDTO) error
}

type GormMessageRepository struct {
	db *gorm.DB
}

func NewGormMessageRepository(db *gorm.DB) *GormMessageRepository {
	return &GormMessageRepository{db: db}
}

func (r *GormMessageRepository) Save(
	ctx context.Context,
	message dto.ChatMessageDTO,
) error {
	record := model.ChatMessage{
		ID:      message.ID,
		RoomID:  message.RoomID,
		UserID:  message.UserID,
		Type:    message.Type,
		Message: message.Message,
		SentAt:  message.SentAt,
	}

	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("save chat message: %w", err)
	}

	return nil
}
