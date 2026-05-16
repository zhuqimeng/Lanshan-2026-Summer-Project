package User

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username  string `json:"username" gorm:"size:64;unique;not null;comment:用户名"`
	Password  string `json:"password" gorm:"size:128;not null;comment:密码哈希"`
	AvatarURL string `gorm:"type:varchar(500)"`
}

type CreateUserReq struct {
	Name     string `json:"name" binding:"required,min=2,max=20"`
	Password string `json:"password" binding:"required,min=6,max=36"`
}

type UpdPwdReq struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=36"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=36"`
}
