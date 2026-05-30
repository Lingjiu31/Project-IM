package hub

import (
	"Project-IM/internal/domain"
	"Project-IM/internal/repository"
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const msgChannel = "im:messages"

type Hub struct {
	clients    map[int64]*Client            // 在线用户表
	groups     map[int64]map[int64]bool     // 群成员表(群ID -> 用户ID -> 是否在线)
	register   chan *Client                 // 注册
	unregister chan *Client                 // 注销
	joinGroup  chan *GroupAction            // 加入群
	leaveGroup chan *GroupAction            // 离开群
	broadcast  chan *domain.WSMessage       // 转发消息
	msgRepo    repository.MessageRepository // 消息存储
	rdb        *redis.Client                // Redis 客户端
	localRoute chan *domain.Message         // 来自 Redis 的消息，只做本地路由
}

type GroupAction struct {
	groupID int64
	userID  int64
}

func NewHub(msgRepo repository.MessageRepository, rdb *redis.Client) *Hub {
	return &Hub{
		clients:    make(map[int64]*Client),
		groups:     make(map[int64]map[int64]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		joinGroup:  make(chan *GroupAction),
		leaveGroup: make(chan *GroupAction),
		broadcast:  make(chan *domain.WSMessage),
		msgRepo:    msgRepo,
		rdb:        rdb,
		localRoute: make(chan *domain.Message, 256),
	}
}

func (h *Hub) UserJoinGroup(userID, groupID int64) {
	h.joinGroup <- &GroupAction{groupID, userID}
}

func (h *Hub) SubscribeRedis(ctx context.Context) {
	// 先订阅
	sub := h.rdb.Subscribe(ctx, msgChannel)
	defer sub.Close()
	ch := sub.Channel()
	for {
		select {
		case redisMsg := <-ch:
			var msg domain.Message
			if err := json.Unmarshal([]byte(redisMsg.Payload), &msg); err != nil {
				zap.L().Error("Redis消息反序列化失败", zap.Error(err))
				continue
			}
			h.localRoute <- &msg
		case <-ctx.Done():
			return
		}
	}
}

// Run 是 Hub 的核心调度循环，所有对 clients/groups 的读写都在这一个 goroutine 里完成
// 用 channel 传递操作请求，避免并发读写 map 导致的 data race
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userID] = client
		case client := <-h.unregister:
			delete(h.clients, client.userID)
		case action := <-h.joinGroup:
			if h.groups[action.groupID] == nil {
				h.groups[action.groupID] = make(map[int64]bool)
			}
			h.groups[action.groupID][action.userID] = true
		case action := <-h.leaveGroup:
			delete(h.groups[action.groupID], action.userID)
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
			// 数据库操作后数据就自动回写
			data, _ := json.Marshal(record)
			if err := h.rdb.Publish(context.Background(), msgChannel, data).Err(); err != nil {
				zap.L().Error("Redis发布消息失败", zap.Error(err))
			}
		case msg := <-h.localRoute:
			// 来自 Redis 的消息，只做本地路由，不存库不发 Redis
			if msg.TargetType == domain.TargetTypeUser {
				// 单聊：直接找接收方
				if target, ok := h.clients[msg.TargetID]; ok {
					data, _ := json.Marshal(msg)
					target.send <- data
				}
			} else if msg.TargetType == domain.TargetTypeGroup {
				// 群聊：遍历群成员，找到在线的就发
				data, _ := json.Marshal(msg)
				for uid := range h.groups[msg.TargetID] {
					if target, ok := h.clients[uid]; ok {
						target.send <- data
					}
				}
			}
		}
	}
}
