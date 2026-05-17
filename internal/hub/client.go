package hub

import (
	"Project-IM/internal/domain"
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Client struct {
	userID int64
	conn   *websocket.Conn
	hub    *Hub
	send   chan []byte
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
	// 出错,注销用户
	c.Unregister()
}

func (c *Client) WritePump() {
	for msg := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
	c.Unregister()
}
