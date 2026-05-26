package domain

import "time"

// User 用户领域模型
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`      // 不序列化到JSON
	Avatar   string `json:"avatar"` // 头像
}

// Message 消息领域模型
type Message struct {
	ID         int64      `json:"id"`          // 消息 ID
	SenderID   int64      `json:"sender_id"`   // 发送者 ID
	TargetID   int64      `json:"target_id"`   // 单聊：接收方UserID；群聊：RoomID
	TargetType TargetType `json:"target_type"` // 消息目标类型(群聊,单聊)
	Content    string     `json:"content"`
	CreatedAt  time.Time  `json:"created_at"`
	Status     MsgStatus  `json:"status"` // 是否已读
}

type TargetType int8

const (
	TargetTypeUser  TargetType = 1 // 单聊
	TargetTypeGroup TargetType = 2 // 群聊
)

type MsgStatus int8

const (
	MsgStatusUnread MsgStatus = 0
	MsgStatusRead   MsgStatus = 1
)

// Group 聊天室（群聊）
type Group struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GroupMember 群成员关系
type GroupMember struct {
	RoomID   int64     `json:"room_id"`
	UserID   int64     `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

// WSMessage WebSocket 传输的消息结构
// 这是客户端和服务端之间的通信协议
type WSMessage struct {
	Type       WSMsgType  `json:"type"`
	SenderID   int64      `json:"sender_id"`
	TargetID   int64      `json:"target_id"`
	TargetType TargetType `json:"target_type"`
	Content    string     `json:"content"`
	Timestamp  int64      `json:"timestamp"`
}

type WSMsgType string

const (
	WSMsgTypeChat   WSMsgType = "chat"   // 聊天消息
	WSMsgTypeSystem WSMsgType = "system" // 系统通知
	WSMsgTypePing   WSMsgType = "ping"   // 心跳
	WSMsgTypePong   WSMsgType = "pong"   // 心跳响应
)
