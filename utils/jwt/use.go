package jwt

import (
	"LanshanSummerProject/app/api/configs"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateTokenPair(userID uint) (*TokenPair, error) {
	// Access Token
	accessClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(JwtSecret)
	if err != nil {
		return nil, err
	}

	// Refresh Token (同样使用 Claims，但过期时间更长)
	refreshClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(JwtSecret)
	if err != nil {
		return nil, err
	}

	// 将 Refresh Token 存储到 Redis（key: refresh:{userID}，value: refreshStr）
	err = storeRefreshToken(userID, refreshStr)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}

func parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func Logout(userID uint, accessTokenString string) error {
	ctx := context.Background()
	// 1. 删除 refresh token
	configs.Cli.Del(ctx, fmt.Sprintf("refresh:%d", userID))

	// 2. 将 access token 加入黑名单（剩余有效期）
	claims, err := parseToken(accessTokenString)
	if err != nil {
		return err
	}
	exp := claims.ExpiresAt.Time
	ttl := time.Until(exp)
	if ttl > 0 {
		// key = blacklist:{accessToken的jti或签名hash}
		err = configs.Cli.Set(ctx, fmt.Sprintf("blacklist:%s", accessTokenString), "1", ttl).Err()
	}
	return err
}
