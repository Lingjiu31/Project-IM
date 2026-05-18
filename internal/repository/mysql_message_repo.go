package repository

import (
	"Project-IM/internal/domain"

	"gorm.io/gorm"
)

type MySQLMessageRepo struct {
	db *gorm.DB
}

func NewMySQLMessageRepo(db *gorm.DB) *MySQLMessageRepo {
	return &MySQLMessageRepo{db: db}
}

func (r *MySQLMessageRepo) InitTable() error {
	return r.db.AutoMigrate(&MessagePO{})
}

func (r *MySQLMessageRepo) Save(msg *domain.Message) error {
	po := toMessagePO(msg)
	// 把收到的消息存入并且更新消息 ID
	if err := r.db.Create(&po).Error; err != nil {
		return err
	}
	msg.ID = po.ID
	return nil
}

func (r *MySQLMessageRepo) FindByUser(senderID, targetID int64, limit, offset int) ([]*domain.Message, error) {
	var pos []*MessagePO
	// 查询互发的消息, 按照时间排序, 确定多少条到多少条
	if err := r.db.Where(
		"(sender_id = ? AND target_id = ?) OR (sender_id = ? AND target_id = ?)",
		senderID, targetID, targetID, senderID,
	).Order("create_at DESC").Limit(limit).Offset(offset).Find(&pos).Error; err != nil {
		return nil, err
	}
	var msgs []*domain.Message
	for _, po := range pos {
		msgs = append(msgs, toDomainMessage(po))
	}
	return msgs, nil
}

func (r *MySQLMessageRepo) FindUnread(userID int64) ([]*domain.Message, error) {
	var pos []*MessagePO
	// 查询未收消息,
	if err := r.db.Where(
		"target_id = ? AND target_type = ? AND status = ?",
		userID, domain.TargetTypeUser, domain.MsgStatusRead,
	).Find(&pos).Error; err != nil {
		return nil, err
	}
	var msgs []*domain.Message
	for _, po := range pos {
		msgs = append(msgs, toDomainMessage(po))
	}
	return msgs, nil
}
