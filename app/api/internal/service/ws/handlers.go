package ws

import (
	"LanshanSummerProject/app/api/configs"
	model "LanshanSummerProject/app/api/internal/model/user"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境需严格校验
	},
}

var hub *Hub

func InitWebSocket(h *Hub) {
	hub = h
}

func WebSocketHandler(c *gin.Context) {
	// 从JWT中获取用户信息（假设 AuthMiddleware 已将 userID 存入 context）
	userID := c.GetUint("userID")

	// 获取用户名（可从数据库查询，或从token中解析）
	var user model.User
	if err := configs.Db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户不存在"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		configs.Sugar.Error("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   user.ID,
		Username: user.Username,
	}
	hub.Register <- client

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}
