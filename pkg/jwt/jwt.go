package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	jwt.RegisteredClaims
	UserID int64
}

type Manager struct {
	secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

func (m *Manager) Generate(userID int64) (string, error) {
	claims := &UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Parse(tokenStr string) (*UserClaims, error) {
	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return m.secret, nil
		},
	)
	if err != nil || !token.Valid {
		return nil, errors.New("token 无效")
	}
	return claims, nil
}
