package router

import (
	"Project-IM/internal/handler"
	"Project-IM/internal/middleware"
	jwtpkg "Project-IM/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func NewRouter(ws *handler.Handler, user *handler.UserHandler, jwtMgr *jwtpkg.Manager) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/register", user.Register)
		api.POST("/login", user.Login)
	}

	// 需要鉴权的路由加中间件
	auth := r.Group("/", middleware.JWTAuth(jwtMgr))
	auth.GET("/ws", ws.ServeWS)

	return r
}
