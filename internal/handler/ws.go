package handler

import (
	"Project-IM/internal/hub"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket 升级失败"})
		return
	}

	userID := c.GetInt64("user_id")
	if userID == 0 {
		conn.Close()
		return
	}
	// 实例化 client
	client := hub.NewClient(userID, conn, h.hub)

	// 启动 goroutine
	client.Register()
	go h.hub.SendOfflineMessage(client)
	go client.ReadPump()
	go client.WritePump()
}
