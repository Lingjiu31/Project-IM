package repository

import (
	"Project-IM/internal/domain"
	"context"
	"time"
)

type MessageRepository interface {
	Save(ctx context.Context, msg *domain.Message) error
	FindByUser(ctx context.Context, senderID, targetID int64, limit, offset int) ([]*domain.Message, error)
	FindUnread(ctx context.Context, userID int64) ([]*domain.Message, error)
	MarkRead(ctx context.Context, msgIDs []int64) error
	FindGroupMessagesSince(ctx context.Context, groupID int64, since time.Time) ([]*domain.Message, error)
}

// limit 表示每次取多少条, offset 表示从第几条开始
