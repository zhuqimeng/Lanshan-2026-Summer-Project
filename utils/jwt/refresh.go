package jwt

import (
	"LanshanSummerProject/app/api/configs"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func validateRefreshToken(userID uint, tokenString string) bool {
	ctx := context.Background()
	stored, err := configs.Cli.Get(ctx, fmt.Sprintf("refresh:%d", userID)).Result()
	if err != nil {
		return false
	}
	return stored == tokenString
}

// 存储 refresh token（用户登录或刷新成功时覆盖旧值）
func storeRefreshToken(userID uint, refreshToken string) error {
	ctx := context.Background()
	return configs.Cli.Set(ctx, fmt.Sprintf("refresh:%d", userID), refreshToken, RefreshTokenExp).Err()
}

func RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return JwtSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	userID := claims.UserID
	storedToken, err := configs.Cli.Get(context.Background(), fmt.Sprintf("refresh:%d", userID)).Result()
	if err != nil || storedToken != req.RefreshToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	newToken, err := GenerateTokenPair(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  newToken.AccessToken,
		"refresh_token": newToken.RefreshToken,
	})
}
