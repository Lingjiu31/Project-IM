package ws

import (
	"Project-IM/internal/hub"
	"Project-IM/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub      *hub.Hub
	upgrader websocket.Upgrader
}

func NewHandler(hub *hub.Hub) *Handler {
	return &Handler{
		hub: hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 开发阶段,允许所有来源
			},
		},
	}
}

// ServeWS 建立连接
func (h *Handler) ServeWS(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	tokenStr := c.Query("token")
	if tokenStr == "" {
		// token 不存在
		return
	}
	claims, err := middleware.ParseToken(tokenStr)
	if err != nil || claims == nil {
		conn.Close()
		return
	}
	// 实例化 client
	client := hub.NewClient(claims.UserID, conn, h.hub)

	// 启动 goroutine
	client.Register()
	go client.ReadPump()
	go client.WritePump()
}
