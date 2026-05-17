package hub

import (
	"Project-IM/internal/domain"
	"encoding/json"
)

type Hub struct {
	clients    map[int64]*Client        // 在线用户表
	rooms      map[int64]map[int64]bool // 群成员表(群ID,用户ID,是否在群)
	register   chan *Client             // 注册
	unregister chan *Client             // 注销
	joinRoom   chan *RoomAction         // 加入群
	leaveRoom  chan *RoomAction         // 离开群
	broadcast  chan *domain.WSMessage   // 转发消息
}

type RoomAction struct {
	roomID int64
	client *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]*Client),
		rooms:      make(map[int64]map[int64]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		joinRoom:   make(chan *RoomAction),
		leaveRoom:  make(chan *RoomAction),
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
		case action := <-h.joinRoom:
			if _, ok := h.rooms[action.roomID][action.client.userID]; !ok {
				h.rooms[action.roomID][action.client.userID] = true
			} else {
				h.rooms[action.roomID] = make(map[int64]bool)
				h.rooms[action.roomID][action.client.userID] = true
			}
		case action := <-h.leaveRoom:
			delete(h.rooms[action.roomID], action.client.userID)
		case msg := <-h.broadcast:
			if msg.TargetType == domain.TargetTypeUser {
				// 单聊
				if target, ok := h.clients[msg.TargetID]; ok {
					data, _ := json.Marshal(msg)
					target.send <- data
				}
			} else if msg.TargetType == domain.TargetTypeRoom {
				// 群聊：遍历 rooms[msg.TargetID]，找到在线的就发
				data, _ := json.Marshal(msg)
				for uid := range h.rooms[msg.TargetID] {
					if target, ok := h.clients[uid]; ok {
						target.send <- data
					}
				}
			}
		}
	}
}
