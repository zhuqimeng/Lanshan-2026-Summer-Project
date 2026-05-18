package router

import (
	"LanshanSummerProject/app/api/configs"
	"LanshanSummerProject/app/api/internal/middleware"
	"LanshanSummerProject/app/api/internal/service/user/info"
	"LanshanSummerProject/utils/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Start() {
	r := gin.Default()
	r.POST("/register", info.Register)
	r.POST("/login", info.Login)
	r.POST("/refresh", jwt.RefreshToken)
	auth := r.Group("/api")
	auth.Use(middleware.Authorize())
	{
		auth.DELETE("logout", info.Logout)
		auth.POST("/avatar/upload", info.UploadAvatar)
	}
	if err := r.Run(":8080"); err != nil {
		configs.Logger.Fatal("Gin run error", zap.Error(err))
	}
}
