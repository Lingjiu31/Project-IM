package middleware

import "github.com/golang-jwt/jwt/v5"

type UserClaims struct {
	jwt.RegisteredClaims
	UserID int64
}

var JWTKey = []byte("im-secret-key-1018")

func ParseToken(tokenStr string) (*UserClaims, error) {
	claims := &UserClaims{}
	// 解析 token
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return JWTKey, nil
		},
	)
	if err != nil || !token.Valid {
		// 解析失败, token 不对, 返回错误在服务层处理
		return nil, err
	}
	return claims, nil
}
