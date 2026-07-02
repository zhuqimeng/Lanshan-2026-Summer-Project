package mychat

import "time"

// ChatType 聊天类型
const (
	ChatSingle = "single" // 单聊
	ChatGroup  = "group"  // 群聊
	ChatBroad  = "broad"  // 广播
)

// Message 通用消息结构（用于客户端与服务器交互）
type Message struct {
	// 基础字段
	ChatType    string    `json:"chat_type"` // single, group, broad
	FromUserID  uint      `json:"from_user_id"`
	ToID        uint      `json:"to_id"`        // 对方用户ID（单聊）或群组ID（群聊）
	MessageType string    `json:"message_type"` // text, image, file, voice
	Content     string    `json:"content"`      // 文本内容 或 文件URL
	CreateTime  time.Time `json:"create_time"`
}

// MessageRecord 用于数据库存储的消息记录（可选）
type MessageRecord struct {
	ID          uint      `gorm:"primarykey"`
	ChatType    string    `gorm:"index;size:20"`
	FromUserID  uint      `gorm:"index"`
	ToID        uint      `gorm:"index"` // 单聊时为对方user_id，群聊时为group_id
	MessageType string    `gorm:"size:20"`
	Content     string    `gorm:"type:text"`
	IsRead      bool      `gorm:"default:false;index"` // 仅单聊有效
	CreatedAt   time.Time `gorm:"index"`
}

type MessageWithSender struct {
	MessageRecord
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}
