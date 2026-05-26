package ws

import (
	"LanshanSummerProject/app/api/configs"
	"LanshanSummerProject/app/api/internal/model/mychat"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SaveMessage(msg mychat.Message, isReceiverOnline bool) error {
	record := mychat.MessageRecord{
		ChatType:    msg.ChatType,
		FromUserID:  msg.FromUserID,
		ToID:        msg.ToID,
		MessageType: msg.MessageType,
		Content:     msg.Content,
		CreatedAt:   msg.CreateTime,
		IsRead:      false, // 默认未读
	}
	// 单聊且接收者在线，直接标记为已读
	if msg.ChatType == mychat.ChatSingle && isReceiverOnline {
		record.IsRead = true
	}
	// 群聊的 IsRead 字段暂不使用，保持 false

	result := configs.Db.Create(&record)
	return result.Error
}

func GetOfflineMessages(c *gin.Context) {
	userID := c.GetUint("userID")
	var messages []mychat.MessageRecord
	err := configs.Db.Where("to_id = ? AND chat_type = ? AND is_read = ?", userID, mychat.ChatSingle, false).Order("created_at asc").Find(&messages).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": messages})
}

type MarkReadRequest struct {
	MessageIDs []uint `json:"message_ids"` // 如果为空，则标记所有未读
}

// MarkMessagesAsRead 标记单聊消息为已读（可传消息ID列表，或全部）
func MarkMessagesAsRead(c *gin.Context) {
	userID := c.GetUint("userID")
	var req MarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := configs.Db.Model(&mychat.MessageRecord{}).Where("to_id = ? AND chat_type = ? AND is_read = ?", userID, mychat.ChatSingle, false)
	if len(req.MessageIDs) > 0 {
		query = query.Where("id IN ?", req.MessageIDs)
	}
	result := query.Update("is_read", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "标记失败"})
		configs.Logger.Error("MarkMessagesAsRead", zap.Error(result.Error))
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": result.RowsAffected})
}
