package hub

import (
	"Project-IM/internal/domain"
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	userID int64
	conn   *websocket.Conn
	hub    *Hub
	send   chan []byte
	once   sync.Once
}

func NewClient(userID int64, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		userID: userID,
		conn:   conn,
		hub:    hub,
		send:   make(chan []byte, 256),
	}
}

func (c *Client) Register() {
	c.hub.register <- c
}

func (c *Client) Unregister() {
	c.hub.unregister <- c
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
