package service

import (
	"Project-IM/internal/domain"
	"Project-IM/internal/repository"
	jwtpkg "Project-IM/pkg/jwt"
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo   repository.UserRepository
	jwtMgr *jwtpkg.Manager
}

func NewUserService(repo repository.UserRepository, jwtMgr *jwtpkg.Manager) *UserService {
	return &UserService{repo: repo, jwtMgr: jwtMgr}
}

func (s *UserService) Register(ctx context.Context, username, password string) error {
	p, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err = s.repo.Save(ctx, &domain.User{
		Username: username,
		Password: string(p),
	}); err != nil {
		if isUniqueConstraintError(err) {
			return errors.New("用户名已存在")
		}
		return err
	}
	return nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("用户名或密码错误")
		}
		return "", err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("用户名或密码错误")
	}
	// 生成 jwt token
	tokenStr, err := s.jwtMgr.Generate(user.ID)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

// 判断是否为唯一索引冲突
func isUniqueConstraintError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
