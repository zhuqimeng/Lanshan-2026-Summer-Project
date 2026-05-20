package User

import "gorm.io/gorm"

type Friendship struct {
	gorm.Model
	UserID    uint   `gorm:"index;not null"`
	FriendID  uint   `gorm:"index;not null"`
	Status    string `gorm:"type:varchar(20);default:'pending'"` // pending, accepted, rejected, blocked
	Remark    string `gorm:"type:varchar(100)"`                  // 备注
	GroupName string `gorm:"type:varchar(50);default:'默认'"`      // 分组名
}

type AddFriendRequest struct {
	FriendID uint `json:"friend_id" binding:"required"`
}

// HandleRequest 处理好友请求（同意/拒绝）
type HandleRequest struct {
	Action string `json:"action" binding:"required,oneof=accept reject"`
}

type UpdateRemarkRequest struct {
	Remark string `json:"remark" binding:"required,max=100"`
}

type UpdateGroupRequest struct {
	GroupName string `json:"group_name" binding:"required,max=50"`
}
