package hub

import (
	"Project-IM/internal/domain"
	"Project-IM/internal/repository"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	pongWait   = 60 * time.Second    // 等待 Pong 的最长时间
	pingPeriod = (pongWait * 9) / 10 // 发 Ping 间隔，必须 < pongWait，留出网络往返时间
	writeWait  = 10 * time.Second    // 每次写操作的超时
)

type Client struct {
	userID     int64
	conn       *websocket.Conn
	hub        *Hub
	send       chan []byte // 待发送消息队列，WritePump 从这里取消息发出去
	once       sync.Once   // 保证 closeOnce 只执行一次，防止 close(send) 被调用两次 panic
	roomAction *RoomAction
	msgRepo    repository.MessageRepository
}

func NewClient(userID int64, conn *websocket.Conn, hub *Hub, msgRepo repository.MessageRepository) *Client {
	return &Client{
		userID:  userID,
		conn:    conn,
		hub:     hub,
		send:    make(chan []byte, 256),
		msgRepo: msgRepo,
	}
}

func (c *Client) Register() {
	c.hub.register <- c
}

func (c *Client) Unregister() {
	c.hub.unregister <- c
}

func (c *Client) JoinRoom(roomID int64) {
	c.hub.joinRoom <- &RoomAction{
		roomID: roomID,
		client: c,
	}
}

func (c *Client) LeaveRoom(roomID int64) {
	c.hub.leaveRoom <- &RoomAction{
		roomID: roomID,
		client: c,
	}
}

// SendOfflineMessage 用户上线时补发未读消息，发完标记为已读
func (c *Client) SendOfflineMessage() {
	msgs, err := c.msgRepo.FindUnread(context.Background(), c.userID)
	if err != nil {
		zap.L().Error("查询离线消息失败", zap.Int64("userID", c.userID), zap.Error(err))
		return
	}
	var msgIDs []int64
	for _, msg := range msgs {
		data, err := json.Marshal(msg)
		if err != nil {
			zap.L().Error("离线消息序列化失败", zap.Int64("msgID", msg.ID), zap.Error(err))
			continue
		}
		c.send <- data
		msgIDs = append(msgIDs, msg.ID)
	}
	if len(msgIDs) == 0 {
		return
	}
	if err = c.msgRepo.MarkRead(context.Background(), msgIDs); err != nil {
		zap.L().Error("标记离线消息已读失败", zap.Int64("userID", c.userID), zap.Error(err))
	}
}

// ReadPump 负责从 WebSocket 读消息，转发到 Hub 广播
// 同时维护心跳：收到 Pong 就刷新读超时，超时未收到则断开
func (c *Client) ReadPump() {
	defer c.closeOnce()
	// SetReadDeadline: 设置读操作截止时间，超时未收到数据则 ReadMessage() 返回 error
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		zap.L().Error("设置读超时失败", zap.Int64("userID", c.userID), zap.Error(err))
		return
	}
	// SetPongHandler: 注册 Pong 帧回调，每次收到客户端的 Pong 自动触发，刷新截止时间续命
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg domain.WSMessage
		if err = json.Unmarshal(p, &msg); err != nil {
			break
		}
		msg.SenderID = c.userID
		c.hub.broadcast <- &msg
	}
}

// WritePump 负责把 send channel 里的消息写出去，同时定时发 Ping 维持心跳
func (c *Client) WritePump() {
	defer c.conn.Close()
	defer c.closeOnce()

	// ticker.C: 每隔 pingPeriod 自动发送一个时间值，用来触发定时发 Ping
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop() // 函数退出时停止 ticker，释放底层定时器资源

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// send channel 被关闭，说明连接要断了
				return
			}
			// SetWriteDeadline: 设置写操作截止时间，防止对端卡死时 WriteMessage 永久阻塞
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				zap.L().Error("设置写超时失败", zap.Int64("userID", c.userID), zap.Error(err))
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				zap.L().Error("发送消息失败", zap.Int64("userID", c.userID), zap.Error(err))
				return
			}
			var m domain.Message
			if err := json.Unmarshal(msg, &m); err != nil {
				zap.L().Error("消息反序列化失败", zap.Int64("userID", c.userID), zap.Error(err))
				continue
			}
			if err := c.msgRepo.MarkRead(context.Background(), []int64{m.ID}); err != nil {
				zap.L().Error("标记消息已读失败", zap.Int64("userID", c.userID), zap.Error(err))
				continue
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				zap.L().Error("设置写超时失败", zap.Int64("userID", c.userID), zap.Error(err))
				return
			}
			// WriteMessage(PingMessage): 发送 WebSocket Ping 帧，客户端协议层自动回 Pong
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				zap.L().Error("发送心跳失败", zap.Int64("userID", c.userID), zap.Error(err))
				return
			}
		}
	}
}

// closeOnce 保证断线清理只执行一次
// ReadPump 和 WritePump 都会触发，sync.Once 防止 close(send) 重复调用 panic
func (c *Client) closeOnce() {
	c.once.Do(func() {
		c.Unregister()
		close(c.send)
	})
}
