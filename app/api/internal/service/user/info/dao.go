package info

import (
	"LanshanSummerProject/app/api/configs"
	User "LanshanSummerProject/app/api/internal/model/user"
	"errors"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(req *User.CreateUserReq) error {
	var count int64
	if err := configs.Db.Model(&User.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		configs.Logger.Error("CreateUser", zap.Error(err))
		return err
	}
	if count > 0 {
		return errors.New("已存在的用户名")
	}
	// 检查用户是否存在

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		configs.Logger.Error("CreateUser", zap.Error(err))
		return err
	}
	user := User.User{
		Username: req.Username,
		Password: string(hash),
	}

	if err := configs.Db.Create(&user).Error; err != nil {
		configs.Logger.Error("CreateUser", zap.Error(err))
		return err
	}
	return nil
}

func ReadUser(req *User.CreateUserReq) (uint, error) {
	var user User.User
	res := configs.Db.Where("username = ?", req.Username).First(&user)
	if res.Error != nil {
		configs.Logger.Error("ReadUser", zap.Error(res.Error))
		return 0, res.Error
	}
	// 查询用户信息
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return 0, errors.New("密码错误")
	}
	return user.ID, nil
}
