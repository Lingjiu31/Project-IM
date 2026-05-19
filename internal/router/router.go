package router

import (
	"Project-IM/internal/handler"
	"Project-IM/internal/ws"

	"github.com/gin-gonic/gin"
)

func NewRouter(ws *ws.Handler, user *handler.UserHandler) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/register", user.Register)
		api.POST("/login", user.Login)
	}

	r.GET("/ws", ws.ServeWS)

	return r
}
