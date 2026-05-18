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
