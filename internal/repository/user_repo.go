package repository

import (
	"Project-IM/internal/domain"
	"context"
)

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error                         // 注册用
	FindByUsername(ctx context.Context, username string) (*domain.User, error) // 登录用
}
