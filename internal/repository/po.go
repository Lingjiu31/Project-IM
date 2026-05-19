package repository

import (
	"Project-IM/internal/domain"
	"time"
)

type MessagePO struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	SenderID   int64     `gorm:"column:sender_id;not null"`
	TargetID   int64     `gorm:"column:target_id;not null"`
	TargetType int8      `gorm:"column:target_type;not null"`
	Content    string    `gorm:"column:content;type:text;not null"`
	Status     int8      `gorm:"column:status;default:0"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

type UserPO struct {
	ID       int64  `gorm:"primaryKey;autoIncrement"`
	Username string `gorm:"column:username;type:varchar(64);not null;uniqueIndex"`
	Password string `gorm:"column:password;type:varchar(256);not null"`
	Avatar   string `gorm:"column:avatar;type:varchar(256);"`
}

func (MessagePO) TableName() string {
	return "messages"
}
func (UserPO) TableName() string {
	return "users"
}

func toMessagePO(msg *domain.Message) *MessagePO {
	return &MessagePO{
		ID:         msg.ID,
		SenderID:   msg.SenderID,
		TargetID:   msg.TargetID,
		TargetType: int8(msg.TargetType),
		Content:    msg.Content,
		Status:     int8(msg.Status),
		CreatedAt:  msg.CreatedAt,
	}
}

func toDomainMessage(po *MessagePO) *domain.Message {
	return &domain.Message{
		ID:         po.ID,
		SenderID:   po.SenderID,
		TargetID:   po.TargetID,
		TargetType: domain.TargetType(po.TargetType),
		Content:    po.Content,
		Status:     domain.MsgStatus(po.Status),
		CreatedAt:  po.CreatedAt,
	}
}

func toDomainUser(po *UserPO) *domain.User {
	return &domain.User{
		ID:       po.ID,
		Username: po.Username,
		Password: po.Password,
		Avatar:   po.Avatar,
	}
}

func toUserPO(user *domain.User) *UserPO {
	return &UserPO{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		Avatar:   user.Avatar,
	}
}
