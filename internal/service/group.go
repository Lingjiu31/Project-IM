package service

import (
	"Project-IM/internal/domain"
	"Project-IM/internal/repository"
	"context"
)

type GroupService struct {
	repo repository.GroupRepository
}

func NewGroupService(repo repository.GroupRepository) *GroupService {
	return &GroupService{repo: repo}
}

func (s *GroupService) CreateGroup(ctx context.Context, name string, creatorID int64,
) (*domain.Group, error) {
	group := &domain.Group{
		Name:      name,
		CreatedBy: creatorID,
	}
	if err := s.repo.CreateGroup(ctx, group); err != nil {
		return nil, err
	}
	if err := s.repo.AddMember(ctx, group.ID, group.CreatedBy); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) JoinGroup(ctx context.Context, groupID int64, userID int64) error {
	if _, err := s.repo.FindByGroupID(ctx, groupID); err != nil {
		return err
	}
	return s.repo.AddMember(ctx, groupID, userID)
}
