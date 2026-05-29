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
	hub       *hub.Hub
	msgRepo   repository.MessageRepository
	groupRepo repository.GroupRepository
	upgrader  websocket.Upgrader
}

func NewHandler(hub *hub.Hub, msgRepo repository.MessageRepository,
	groupRepo repository.GroupRepository) *Handler {
	return &Handler{
		hub:       hub,
		msgRepo:   msgRepo,
		groupRepo: groupRepo,
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

	client := hub.NewClient(userID, conn, h.hub, h.msgRepo, h.groupRepo)
	client.Register()
	members, err := h.groupRepo.FindGroupsByUserID(c.Request.Context(), userID)
	if err != nil {
		zap.L().Error("查询群聊失败", zap.Int64("userID", userID), zap.Error(err))
	}
	for _, member := range members {
		client.JoinGroup(member.GroupID)
	}
	go client.SendOfflineMessage()
	go client.SendGroupOfflineMessage()
	go client.ReadPump()
	go client.WritePump()
}
