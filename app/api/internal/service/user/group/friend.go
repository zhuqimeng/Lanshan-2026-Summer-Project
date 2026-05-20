package group

import (
	"LanshanSummerProject/app/api/configs"
	"net/http"
	"strconv"

	model "LanshanSummerProject/app/api/internal/model/user"

	"github.com/gin-gonic/gin"
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

func HandleFriendRequest(c *gin.Context) {
	currentUserID := c.GetUint("userID")

	requestID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求ID"})
		return
	}

	var req model.HandleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 查找待处理的请求（必须是当前用户作为被请求方）
	var friendship model.Friendship
	err = configs.Db.Where("id = ? AND friend_id = ? AND status = ?", requestID, currentUserID, "pending").First(&friendship).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "请求不存在或已处理"})
		return
	}

	if req.Action == "accept" {
		// 更新原请求状态为 accepted
		friendship.Status = "accepted"
		if err := configs.Db.Save(&friendship).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "接受失败"})
			return
		}
		// 创建反向关系记录（表示对方也添加你为好友）
		reverse := model.Friendship{
			UserID:    currentUserID,
			FriendID:  friendship.UserID,
			Status:    "accepted",
			Remark:    "",
			GroupName: "默认",
		}
		if err := configs.Db.Create(&reverse).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "建立好友关系失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已接受好友请求"})
	} else { // reject
		if err := configs.Db.Delete(&friendship).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "拒绝失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已拒绝好友请求"})
	}
}

// GetFriendRequests 获取好友请求列表（待处理）
func GetFriendRequests(c *gin.Context) {
	currentUserID := c.GetUint("userID")

	var requests []model.Friendship
	err := configs.Db.Where("friend_id = ? AND status = ?", currentUserID, "pending").Find(&requests).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 构建返回数据（包含发送者的基本信息）
	type RequestInfo struct {
		ID        uint   `json:"id"`
		UserID    uint   `json:"user_id"`
		Username  string `json:"username"`
		Avatar    string `json:"avatar"`
		CreatedAt string `json:"created_at"`
	}
	result := make([]RequestInfo, 0)
	for _, r := range requests {
		var user model.User
		configs.Db.First(&user, r.UserID) // 依照主键查询库中元素
		result = append(result, RequestInfo{
			ID:        r.ID,
			UserID:    user.ID,
			Username:  user.Username,
			Avatar:    user.AvatarURL,
			CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, result)
}

func DeleteFriend(c *gin.Context) {
	currentUserID := c.GetUint("userID")

	friendID, err := strconv.Atoi(c.Param("friend_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的好友ID"})
		return
	}

	// 删除双向记录
	err = configs.Db.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		currentUserID, friendID, friendID, currentUserID).
		Delete(&model.Friendship{}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除好友失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
