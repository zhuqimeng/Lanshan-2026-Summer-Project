package ws

import (
	"LanshanSummerProject/app/api/configs"
	"LanshanSummerProject/app/api/internal/model/mychat"
	"encoding/json"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   uint
	Username string
}

// ReadPump 读取客户端消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// 解析客户端发来的 JSON 消息
		var msg mychat.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			configs.Logger.Warn("解析消息错误:", zap.Error(err))
			continue
		}
		msg.FromUserID = c.UserID
		// 交给 Hub 处理
		c.Hub.Route <- msg
	}
}

// WritePump 向客户端写入消息
func (c *Client) WritePump() {
	defer c.Conn.Close()
	for msgBytes := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msgBytes)
		if err != nil {
			break
		}
	}
}
