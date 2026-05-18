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

func (MessagePO) TableName() string {
	return "messages"
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
