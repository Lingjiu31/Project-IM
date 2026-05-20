package handler

import (
	"Project-IM/internal/hub"
	jwtpkg "Project-IM/pkg/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub      *hub.Hub
	upgrader websocket.Upgrader
	jwtMgr   *jwtpkg.Manager
}

func NewHandler(hub *hub.Hub, jwtMgr *jwtpkg.Manager) *Handler {
	return &Handler{
		hub:    hub,
		jwtMgr: jwtMgr,
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

	tokenStr := c.Query("token")
	if tokenStr == "" {
		// token 不存在
		conn.Close()
		return
	}
	claims, err := h.jwtMgr.Parse(tokenStr)
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
