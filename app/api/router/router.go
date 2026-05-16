package router

import (
	"LanshanSummerProject/app/api/configs"
	"LanshanSummerProject/app/api/internal/service/user/info"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Start() {
	r := gin.Default()
	r.POST("/register", info.Register)
	r.POST("/login", info.Login)
	if err := r.Run(":8080"); err != nil {
		configs.Logger.Fatal("Gin run error", zap.Error(err))
	}
}
