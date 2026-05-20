package group

import (
	"LanshanSummerProject/app/api/configs"
	model "LanshanSummerProject/app/api/internal/model/user"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetFriendList(c *gin.Context) {
	currentUserID := c.GetUint("userID")

	group := c.Query("group") // 分组名，可选

	query := configs.Db.Where("user_id = ? AND status = ?", currentUserID, "accepted")
	if group != "" {
		query = query.Where("group_name = ?", group)
	}

	var friendships []model.Friendship
	if err := query.Find(&friendships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 收集好友ID
	friendIDs := make([]uint, len(friendships))
	for i, f := range friendships {
		friendIDs[i] = f.FriendID
	}

	var friends []model.User
	configs.Db.Where("id IN ?", friendIDs).Find(&friends)

	friendMap := make(map[uint]model.User)
	for _, f := range friends {
		friendMap[f.ID] = f
	}

	type FriendInfo struct {
		ID        uint   `json:"id"`
		Username  string `json:"username"`
		Avatar    string `json:"avatar"`
		Remark    string `json:"remark"`
		GroupName string `json:"group_name"`
	}
	result := make([]FriendInfo, 0)
	for _, f := range friendships {
		user := friendMap[f.FriendID]
		result = append(result, FriendInfo{
			ID:        user.ID,
			Username:  user.Username,
			Avatar:    user.AvatarURL,
			Remark:    f.Remark,
			GroupName: f.GroupName,
		})
	}
	c.JSON(http.StatusOK, result)
}

func UpdateFriendRemark(c *gin.Context) {
	currentUserID := c.GetUint("userID")

	friendID, err := strconv.Atoi(c.Param("friend_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的好友ID"})
		return
	}

	var req model.UpdateRemarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	result := configs.Db.Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ? AND status = ?", currentUserID, friendID, "accepted").
		Update("remark", req.Remark)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "好友关系不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "备注更新成功"})
}

func UpdateFriendGroup(c *gin.Context) {
	currentUserID := c.GetUint("userID")

	friendID, err := strconv.Atoi(c.Param("friend_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的好友ID"})
		return
	}

	var req model.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	result := configs.Db.Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ? AND status = ?", currentUserID, friendID, "accepted").
		Update("group_name", req.GroupName)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "好友关系不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "分组更新成功"})
}

func GetFriendGroups(c *gin.Context) {
	currentUserID := c.GetUint("userID")
	var groups []string
	configs.Db.Model(&model.Friendship{}).
		Where("user_id = ? AND status = ?", currentUserID, "accepted").
		Distinct("group_name").
		Pluck("group_name", &groups)
	c.JSON(http.StatusOK, groups)
}
