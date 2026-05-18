package group

import (
	"LanshanSummerProject/app/api/configs"
	"net/http"

	"github.com/gin-gonic/gin"

	model "LanshanSummerProject/app/api/internal/model/user"
)

func SendFriendRequest(c *gin.Context) {
	currentUserID := c.GetUint("userID")
	var req model.AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if req.FriendID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能添加自己为好友"})
		return
	}
	var existing model.Friendship
	err := configs.Db.Where("user_id = ? AND friend_id = ?", currentUserID, req.FriendID).First(&existing).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "已发送过好友请求或已是好友"})
		return
	}
	// 创建请求记录
	friendship := model.Friendship{
		UserID:    currentUserID,
		FriendID:  req.FriendID,
		Status:    "pending",
		Remark:    "",
		GroupName: "默认",
	}
	if err := configs.Db.Create(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送请求失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "好友请求已发送"})
}
