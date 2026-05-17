package hub

import (
	"Project-IM/internal/domain"
	"encoding/json"
)

type Hub struct {
	clients    map[int64]*Client      // "在线用户表"
	register   chan *Client           // 注册
	unregister chan *Client           // 注销
	broadcast  chan *domain.WSMessage // 转发消息
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *domain.WSMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userID] = client
		case client := <-h.unregister:
			delete(h.clients, client.userID)
		case msg := <-h.broadcast:
			if target, ok := h.clients[msg.TargetID]; ok {
				data, _ := json.Marshal(msg)
				target.send <- data
			}
		}
	}
}
