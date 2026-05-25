package hub

import (
	"Project-IM/internal/domain"
	"Project-IM/internal/repository"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait   = 60 * time.Second    // 等待 Pong 的最长时间
	pingPeriod = (pongWait * 9) / 10 // 发 Ping 间隔
	writeWait  = 10 * time.Second    // 每次写操作的超时
)

type Client struct {
	userID     int64
	conn       *websocket.Conn
	hub        *Hub
	send       chan []byte
	once       sync.Once
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

func (c *Client) SendOfflineMessage() {
	msgs, err := c.msgRepo.FindUnread(context.Background(), c.userID)
	if err != nil {
		log.Println(err)
		return
	}
	var msgIDs []int64
	for _, msg := range msgs {
		data, err := json.Marshal(msg)
		if err != nil {
			log.Println(err)
			// 只跳过这一个
			continue
		}
		c.send <- data
		msgIDs = append(msgIDs, msg.ID)
	}
	if len(msgIDs) == 0 {
		return
	}
	if err = c.msgRepo.MarkRead(context.Background(), msgIDs); err != nil {
		log.Println(err)
		return
	}
}

func (c *Client) ReadPump() {
	// 不管因为什么原因出错结束函数,都会关闭通信通道
	defer c.closeOnce()
	// SetReadDeadline: 设置读操作截止时间，超时未收到数据则 ReadMessage() 返回 error
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}
	// SetPongHandler: 注册 Pong 帧回调，每次收到客户端的 Pong 自动触发
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// 把 p 解析到 msg 里
		var msg domain.WSMessage
		err = json.Unmarshal(p, &msg)
		if err != nil {
			break
		}
		msg.SenderID = c.userID
		c.hub.broadcast <- &msg
	}
}

func (c *Client) WritePump() {
	defer c.conn.Close()
	defer c.closeOnce()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			// SetWriteDeadline: 设置写操作截止时间，防止对端卡死时 WriteMessage 永久阻塞
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Println(err)
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println(err)
				return
			}
			var m domain.Message
			if err := json.Unmarshal(msg, &m); err != nil {
				log.Println(err)
				continue
			}
			if err := c.msgRepo.MarkRead(context.Background(), []int64{m.ID}); err != nil {
				log.Println(err)
				continue
			}
		case <-ticker.C:
			// ticker.C: 每隔 pingPeriod 自动发送一个时间值，用来触发定时任务
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Println(err)
				return
			}
			// WriteMessage(PingMessage): 发送 WebSocket Ping 帧，客户端协议层自动回 Pong
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func (c *Client) closeOnce() {
	c.once.Do(func() {
		// 注销用户
		c.Unregister()
		// 删除传送带
		close(c.send)
	})
}
