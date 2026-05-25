package router

import (
	"Project-IM/internal/handler"
	"Project-IM/internal/middleware"
	jwtpkg "Project-IM/pkg/jwt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(ws *handler.Handler, user *handler.UserHandler, jwtMgr *jwtpkg.Manager) *gin.Engine {
	r := gin.Default()

	// CORS 中间件，开发阶段允许所有来源
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

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
