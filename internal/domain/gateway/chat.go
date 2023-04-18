package gateway

import (
	"context"

	"github.com/aureanemoraes/chatservice/internal/domain/entity"
)

type ChatGateway interface {
	// instance entity
	CreateChat(ctx context.Context, chat *entity.Chat) error
	// find on db
	FindChatByID(ctx context.Context, chatID string) (*entity.Chat, error)
	// save on db
	SaveChat(ctx context.Context, chat *entity.Chat) error
}
