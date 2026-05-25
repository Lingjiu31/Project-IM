package handler

import (
	"Project-IM/internal/hub"
	"Project-IM/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Handler struct {
	hub      *hub.Hub
	msgRepo  repository.MessageRepository
	upgrader websocket.Upgrader
}

func NewHandler(hub *hub.Hub, msgRepo repository.MessageRepository) *Handler {
	return &Handler{
		hub:     hub,
		msgRepo: msgRepo,
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
		zap.L().Warn("WebSocket 连接缺少 user_id，拒绝连接")
		conn.Close()
		return
	}

	client := hub.NewClient(userID, conn, h.hub, h.msgRepo)
	client.Register()
	go client.SendOfflineMessage()
	go client.ReadPump()
	go client.WritePump()
}
