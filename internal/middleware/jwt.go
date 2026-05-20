package middleware

import (
	jwtpkg "Project-IM/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth(jwtMgr *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token 无效"})
			c.Abort() // 终止
			return
		}
		tokenStr := parts[1]
		claims, err := jwtMgr.Parse(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token 无效"})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
