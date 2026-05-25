package middleware

import (
	jwtpkg "Project-IM/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth(jwtMgr *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先正常从 Header 取, 没有就从 query 取(WebSocket 测试场景)
		authHeader := c.GetHeader("Authorization")
		var tokenStr string
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token 无效"})
				c.Abort() // 终止
				return
			}
			tokenStr = parts[1]
		} else {
			tokenStr = c.Query("token")
		}
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token 无效"})
			c.Abort()
			return
		}
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
