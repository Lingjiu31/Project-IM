package hub

import (
	"Project-IM/internal/domain"
	"Project-IM/internal/repository"
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
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
	for msg := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
		var m domain.Message
		if err = json.Unmarshal(msg, &m); err != nil {
			log.Println(err)
			continue
		}
		if err = c.msgRepo.MarkRead(context.Background(), []int64{m.ID}); err != nil {
			log.Println(err)
			continue
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
