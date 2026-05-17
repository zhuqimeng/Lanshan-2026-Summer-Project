package jwt

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

var (
	JwtSecret []byte
)

const (
	AccessTokenExp  = 15 * time.Minute
	RefreshTokenExp = 7 * 24 * time.Hour
)

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

func init() {
	viper.SetConfigFile("app/api/configs/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config file not loaded: %v", err)
	}
	viper.AutomaticEnv()
	err := viper.BindEnv("jwt.secret", "JWT_SECRET")
	if err != nil {
		return
	}

	secret := viper.GetString("jwt.secret")
	if secret == "" {
		log.Fatal("JWT_SECRET must be set via config.yaml or environment variable")
	}
	JwtSecret = []byte(secret)
}
