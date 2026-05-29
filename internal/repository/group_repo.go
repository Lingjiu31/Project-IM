package repository

import (
	"Project-IM/internal/domain"
	"context"
	"time"
)

type GroupRepository interface {
	CreateGroup(ctx context.Context, group *domain.Group) error
	FindByGroupID(ctx context.Context, groupID int64) (*domain.Group, error)
	AddMember(ctx context.Context, groupID int64, userID int64) error
	FindMembers(ctx context.Context, groupID int64) ([]*domain.GroupMember, error)
	FindGroupsByUserID(ctx context.Context, userID int64) ([]*domain.GroupMember, error)
	UpdateLastSeen(ctx context.Context, userID, groupID int64, t time.Time) error
}
