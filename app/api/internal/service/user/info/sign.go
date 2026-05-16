package info

import (
	"LanshanSummerProject/app/api/configs"
	User "LanshanSummerProject/app/api/internal/model/user"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Register(c *gin.Context) {
	var req User.CreateUserReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := CreateUser(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
	configs.Logger.Info("register success", zap.String("username", req.Username))
}

func Login(c *gin.Context) {
	var req User.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		configs.Logger.Error("login", zap.Error(err))
		return
	}
	if err := ReadUser(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		configs.Logger.Error("login", zap.Error(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
	configs.Logger.Info("login", zap.String("username", req.Username), zap.String("status", "success"))
}
