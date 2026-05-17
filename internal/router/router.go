package router

import (
	"Project-IM/internal/ws"

	"github.com/gin-gonic/gin"
)

func NewRouter(ws *ws.Handler) *gin.Engine {
	r := gin.Default()

	r.GET("/ws", ws.ServeWS)

	return r
}
