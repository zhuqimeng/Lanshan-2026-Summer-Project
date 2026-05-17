package middleware

import (
	"LanshanSummerProject/app/api/configs"
	Myjwt "LanshanSummerProject/utils/jwt"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(200, gin.H{"error": "Authorization header is empty"})
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(200, gin.H{"error": "Authorization header format is invalid"})
			c.Abort()
			return
		}
		tokenString := parts[1]
		ctx := c.Request.Context()
		exists, err := configs.Cli.Exists(ctx, fmt.Sprintf("blacklist:%s", tokenString)).Result()
		if err == nil && exists == 1 {
			c.JSON(200, gin.H{"error": "token 已失效"})
			c.Abort()
			return
		}
		claims := &Myjwt.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return Myjwt.JwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
