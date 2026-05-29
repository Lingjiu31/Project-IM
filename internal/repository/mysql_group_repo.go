package repository

import (
	"Project-IM/internal/domain"
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MySQLGroupRepo struct {
	db *gorm.DB
}

func (r *MySQLGroupRepo) FindGroupsByUserID(ctx context.Context, userID int64) ([]*domain.GroupMember, error) {
	var pos []GroupMemberPO
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).Find(&pos).Error; err != nil {
		return nil, err
	}
	members := make([]*domain.GroupMember, 0, len(pos))
	for _, po := range pos {
		members = append(members, toDomainGroupMember(&po))
	}
	return members, nil
}

func NewMySQLGroupRepo(db *gorm.DB) GroupRepository {
	return &MySQLGroupRepo{
		db: db,
	}
}

func (r *MySQLGroupRepo) CreateGroup(ctx context.Context, group *domain.Group) error {
	po := toGroupPO(group)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	group.ID = po.ID
	return nil
}

func (r *MySQLGroupRepo) FindByGroupID(ctx context.Context, groupID int64) (*domain.Group, error) {
	var po GroupPO
	if err := r.db.WithContext(ctx).
		First(&po, groupID).Error; err != nil {
		return nil, err
	}
	return toDomainGroup(&po), nil
}

func (r *MySQLGroupRepo) AddMember(ctx context.Context, groupID int64, userID int64) error {
	po := &GroupMemberPO{
		GroupID:  groupID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(po).Error
}

func (r *MySQLGroupRepo) FindMembers(ctx context.Context, groupID int64) ([]*domain.GroupMember, error) {
	var pos []GroupMemberPO
	if err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).Find(&pos).Error; err != nil {
		return nil, err
	}
	members := make([]*domain.GroupMember, 0, len(pos))
	for _, po := range pos {
		members = append(members, toDomainGroupMember(&po))
	}
	return members, nil
}
