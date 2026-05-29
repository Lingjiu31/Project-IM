package domain

import "time"

// Group 群组领域模型
type Group struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"created_by"` // 群主 UserID
	CreatedAt time.Time `json:"created_at"`
}

// GroupMember 群成员关系，GroupID + UserID 唯一确定一条关系
type GroupMember struct {
	GroupID    int64      `json:"group_id"`
	UserID     int64      `json:"user_id"`
	JoinedAt   time.Time  `json:"joined_at"`
	LastSeenAt *time.Time `json:"last_seen_at"`
}
