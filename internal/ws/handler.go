package ws

import (
	"Project-IM/internal/hub"
	"net/http"
	"strconv"

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
	// 取 user_id
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return
	}

	// 实例化 client
	client := hub.NewClient(userID, conn, h.hub)

	// 启动 goroutine
	client.Register()
	go client.ReadPump()
	go client.WritePump()
}
