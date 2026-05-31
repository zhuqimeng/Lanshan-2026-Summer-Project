package mychat

import (
	"time"

	"gorm.io/gorm"
)

type Group struct {
	ID          uint   `gorm:"primarykey"`
	GroupNumber uint   `gorm:"uniqueIndex;not null"`   // 群号，由业务生成（如从100000开始递增）
	Name        string `gorm:"type:varchar(50);index"` // 群名称（允许重复）
	Avatar      string `gorm:"size:200"`               // 群头像
	OwnerID     uint   `gorm:"index"`                  // 群主用户ID
	Intro       string `gorm:"type:varchar(255)"`      // 群简介
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// GroupMember 群成员表（联合唯一索引保证不会重复加群）
type GroupMember struct {
	ID        uint   `gorm:"primarykey"`
	GroupID   uint   `gorm:"index;uniqueIndex:idx_group_user"`
	UserID    uint   `gorm:"index;uniqueIndex:idx_group_user"`
	Role      string `gorm:"type:varchar(20);default:'member'"` // owner, admin, member
	Remark    string `gorm:"type:varchar(100)"`                 // 群内昵称/备注
	JoinedAt  time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
