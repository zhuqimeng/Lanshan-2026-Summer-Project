package info

import (
	"LanshanSummerProject/app/api/configs"
	User "LanshanSummerProject/app/api/internal/model/user"
	Myjwt "LanshanSummerProject/utils/jwt"
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
	userID, err := ReadUser(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		configs.Logger.Error("login", zap.Error(err))
		return
	}
	tokenPair, err := Myjwt.GenerateTokenPair(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		configs.Logger.Error("login", zap.Error(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"token":   tokenPair,
	})
	configs.Logger.Info("login", zap.String("username", req.Username), zap.String("status", "success"))
}
