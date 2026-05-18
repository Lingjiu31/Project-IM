package repository

import "Project-IM/internal/domain"

type MessageRepository interface {
	Save(msg *domain.Message) error
	FindByUser(senderID, targetID int64, limit, offset int) ([]*domain.Message, error)
	FindUnread(userID int64) ([]*domain.Message, error)
}

// limit 表示每次取多少条, offset 表示从第几条开始
