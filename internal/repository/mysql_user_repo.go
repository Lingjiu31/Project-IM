package repository

import (
	"Project-IM/internal/domain"
	"context"

	"gorm.io/gorm"
)

type MySQLUserRepo struct {
	db *gorm.DB
}

func NewMySQLUserRepo(db *gorm.DB) *MySQLUserRepo {
	return &MySQLUserRepo{db: db}
}

func (r *MySQLUserRepo) InitTable() error {
	return r.db.AutoMigrate(&UserPO{})
}

func (r *MySQLUserRepo) Save(ctx context.Context, user *domain.User) error {
	po := toUserPO(user)
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	user.ID = po.ID
	return nil
}

func (r *MySQLUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var po UserPO
	if err := r.db.WithContext(ctx).Where("username = ?", username).
		First(&po).Error; err != nil {
		return nil, err
	}
	return toDomainUser(&po), nil
}
