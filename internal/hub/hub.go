package hub

import (
	"Project-IM/internal/domain"
	"Project-IM/internal/repository"
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

type Hub struct {
	clients    map[int64]*Client            // 在线用户表
	rooms      map[int64]map[int64]bool     // 群成员表(群ID,用户ID,是否在线)
	register   chan *Client                 // 注册
	unregister chan *Client                 // 注销
	joinRoom   chan *RoomAction             // 加入群
	leaveRoom  chan *RoomAction             // 离开群
	broadcast  chan *domain.WSMessage       // 转发消息
	msgRepo    repository.MessageRepository // 消息存储
}

type RoomAction struct {
	roomID int64
	client *Client
}

func NewHub(msgRepo repository.MessageRepository) *Hub {
	return &Hub{
		clients:    make(map[int64]*Client),
		rooms:      make(map[int64]map[int64]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		joinRoom:   make(chan *RoomAction),
		leaveRoom:  make(chan *RoomAction),
		broadcast:  make(chan *domain.WSMessage),
		msgRepo:    msgRepo,
	}
}

// Run 是 Hub 的核心调度循环，所有对 clients/rooms 的读写都在这一个 goroutine 里完成
// 用 channel 传递操作请求，避免并发读写 map 导致的 data race
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userID] = client
		case client := <-h.unregister:
			delete(h.clients, client.userID)
		case action := <-h.joinRoom:
			if h.rooms[action.roomID] == nil {
				// 群不存在则初始化
				h.rooms[action.roomID] = make(map[int64]bool)
			}
			// 不管成员是否已存在直接赋值
			h.rooms[action.roomID][action.client.userID] = true
		case action := <-h.leaveRoom:
			delete(h.rooms[action.roomID], action.client.userID)
		case msg := <-h.broadcast:
			// 先存库，再转发
			record := &domain.Message{
				SenderID:   msg.SenderID,
				TargetID:   msg.TargetID,
				TargetType: msg.TargetType,
				Content:    msg.Content,
				Status:     domain.MsgStatusUnread,
				CreatedAt:  time.Now(),
			}
			if err := h.msgRepo.Save(context.Background(), record); err != nil {
				zap.L().Error("消息存库失败", zap.Error(err))
			}
			if msg.TargetType == domain.TargetTypeUser {
				// 单聊：直接找接收方
				if target, ok := h.clients[msg.TargetID]; ok {
					data, _ := json.Marshal(msg)
					target.send <- data
				}
			} else if msg.TargetType == domain.TargetTypeRoom {
				// 群聊：遍历群成员，找到在线的就发
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
