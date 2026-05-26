package ws

import (
	"LanshanSummerProject/app/api/configs"
	"LanshanSummerProject/app/api/internal/model/mychat"
	"encoding/json"
	"sync"
)

type Hub struct {
	// 所有在线客户端，key: userID
	Clients   map[uint]*Client
	ClientsMu sync.RWMutex
	// 注册,注销通道
	Register   chan *Client
	Unregister chan *Client
	// 消息路由
	Route chan mychat.Message
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uint]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Route:      make(chan mychat.Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.ClientsMu.Lock()
			h.Clients[client.UserID] = client
			h.ClientsMu.Unlock()
			configs.Sugar.Info("用户 %d 上线", client.UserID)

		case client := <-h.Unregister:
			h.ClientsMu.Lock()
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.Send)
				configs.Sugar.Info("用户 %d 下线", client.UserID)
			}
			h.ClientsMu.Unlock()

		case msg := <-h.Route:
			h.handleMessage(msg)
		}
	}
}

// 消息路由逻辑
func (h *Hub) handleMessage(msg mychat.Message) {
	// 判断接收者是否在线（仅单聊需要）
	receiverOnline := false
	if msg.ChatType == mychat.ChatSingle {
		h.ClientsMu.RLock()
		_, ok := h.Clients[msg.ToID]
		h.ClientsMu.RUnlock()
		receiverOnline = ok
	}
	// 保存消息到数据库（异步，避免阻塞消息路由）
	go func() {
		if err := SaveMessage(msg, receiverOnline); err != nil {
			configs.Sugar.Error("保存消息失败: ", err)
		}
	}()
	// 路由消息...
	switch msg.ChatType {
	case mychat.ChatSingle:
		if receiverOnline {
			h.sendToUser(msg.ToID, msg)
		} else {
			// 离线消息已保存，无需额外操作
			configs.Sugar.Info("用户不在线，消息已存为离线，离线用户ID为：", msg.ToID)
		}
	case mychat.ChatGroup:
		h.sendToGroup(msg.ToID, msg)
	case mychat.ChatBroad:
		h.broadcastToAll(msg)
	default:
		configs.Sugar.Warn("未知聊天类型")
	}
}

// 发送给某个用户（单聊）
func (h *Hub) sendToUser(userID uint, msg mychat.Message) {
	h.ClientsMu.RLock()
	client, ok := h.Clients[userID]
	h.ClientsMu.RUnlock()
	if ok {
		data, _ := json.Marshal(msg)
		select {
		case client.Send <- data:
		default:
			// 发送队列满，关闭连接
			close(client.Send)
			delete(h.Clients, userID)
		}
	}
}

// 群聊：遍历所有客户端，查找属于该群组的成员
func (h *Hub) sendToGroup(groupID uint, msg mychat.Message) {
	// 需要从数据库获取群组成员ID列表，此处简化：假设有函数 GetGroupMemberIDs(groupID) []uint
	memberIDs := GetGroupMemberIDs(groupID)
	for _, uid := range memberIDs {
		h.sendToUser(uid, msg)
	}
}

// 广播：给所有在线用户发送消息（不包括自己可选）
func (h *Hub) broadcastToAll(msg mychat.Message) {
	h.ClientsMu.RLock()
	defer h.ClientsMu.RUnlock()
	data, _ := json.Marshal(msg)
	for _, client := range h.Clients {
		// 可根据需要排除消息发送者：if client.UserID == msg.FromUserID { continue }
		select {
		case client.Send <- data:
		default:
		}
	}
}

func GetGroupMemberIDs(groupID uint) []uint {
	return []uint{groupID}
}
