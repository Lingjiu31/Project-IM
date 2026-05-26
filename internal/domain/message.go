package domain

import "time"

// Message 消息领域模型
type Message struct {
	ID         int64      `json:"id"`
	SenderID   int64      `json:"sender_id"`   // 发送者 ID
	TargetID   int64      `json:"target_id"`   // 单聊：接收方 UserID；群聊：GroupID
	TargetType TargetType `json:"target_type"` // 消息目标类型
	Content    string     `json:"content"`
	CreatedAt  time.Time  `json:"created_at"`
	Status     MsgStatus  `json:"status"` // 已读/未读
}

// TargetType 消息目标类型
type TargetType int8

const (
	TargetTypeUser  TargetType = 1 // 单聊
	TargetTypeGroup TargetType = 2 // 群聊
)

// MsgStatus 消息已读状态
type MsgStatus int8

const (
	MsgStatusUnread MsgStatus = 0
	MsgStatusRead   MsgStatus = 1
)

// WSMessage WebSocket 传输的消息结构，是客户端和服务端之间的通信协议
type WSMessage struct {
	SenderID   int64      `json:"sender_id"`
	TargetID   int64      `json:"target_id"`
	TargetType TargetType `json:"target_type"`
	Content    string     `json:"content"`
	Timestamp  int64      `json:"timestamp"`
}
