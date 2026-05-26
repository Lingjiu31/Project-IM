package domain

// User 用户领域模型
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`      // 不序列化到 JSON
	Avatar   string `json:"avatar"` // 头像
}
