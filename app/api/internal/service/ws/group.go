package ws

import (
	"LanshanSummerProject/app/api/configs"
	"LanshanSummerProject/app/api/internal/model/mychat"
	model "LanshanSummerProject/app/api/internal/model/user"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CreateGroupRequest struct {
	Name   string `json:"name" binding:"required,max=50"`
	Intro  string `json:"intro" binding:"max=255"`
	Avatar string `json:"avatar"` // 群头像URL，可选
}

func GenerateGroupNumber() (uint, error) {
	ctx := context.Background()
	num, err := configs.Cli.Incr(ctx, "group_number_counter").Result()
	if err != nil {
		return 0, err
	}
	if num == 1 {
		configs.Cli.Set(ctx, "group_number_counter", 100000, 0)
		return 100000, nil
	}
	return uint(num), nil
}

func CreateGroup(c *gin.Context) {
	userID := c.GetUint("userID")
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	// 创建群组
	groupNumber, err := GenerateGroupNumber()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成群号失败"})
		configs.Logger.Error("生成群号失败", zap.Any("err", err))
		return
	}
	group := mychat.Group{
		GroupNumber: groupNumber,
		Name:        req.Name,
		Intro:       req.Intro,
		Avatar:      req.Avatar,
		OwnerID:     userID,
	}
	if err := configs.Db.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建群聊失败"})
		configs.Logger.Error("创建群聊失败", zap.Any("err", err))
		return
	}
	// 将创建者添加为群主（owner）
	member := mychat.GroupMember{
		GroupID:  group.ID,
		UserID:   userID,
		Role:     "owner",
		JoinedAt: time.Now(),
	}
	if err := configs.Db.Create(&member).Error; err != nil {
		configs.Db.Delete(&group)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加群主失败"})
		configs.Logger.Error("添加群主失败", zap.Any("err", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"group_number": group.GroupNumber,
		"group_name":   group.Name,
		"intro":        group.Intro,
		"avatar":       group.Avatar,
		"created_at":   group.CreatedAt,
	})
}

// JoinGroupRequest 加入群聊（通过群号）
type JoinGroupRequest struct {
	GroupNumber uint   `json:"group_number" binding:"required"`
	Remark      string `json:"remark" binding:"required,max=255"`
}

func JoinGroup(c *gin.Context) {
	userID := c.GetUint("userID")
	var req JoinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var group mychat.Group
	if err := configs.Db.Where("group_number = ?", req.GroupNumber).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "群聊不存在"})
		return
	}
	var existing mychat.GroupMember
	if err := configs.Db.Where("group_id = ? AND user_id = ?", group.ID, userID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "已经是群成员"})
		return
	}
	// 加入群聊
	member := mychat.GroupMember{
		GroupID:  group.ID,
		UserID:   userID,
		Role:     "member",
		Remark:   req.Remark,
		JoinedAt: time.Now(),
	}
	if err := configs.Db.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加入群聊失败"})
		configs.Logger.Error("加入群聊失败", zap.Any("err", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":      "加入成功",
		"group_number": group.GroupNumber,
		"group_name":   group.Name,
	})
}

// LeaveGroup 退出群聊（群主不能直接退出，需要先转让或解散）
func LeaveGroup(c *gin.Context) {
	userID := c.GetUint("userID")
	groupNumber, err := strconv.Atoi(c.Param("group_number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的群号"})
		return
	}
	// 通过群号找到群组
	var group mychat.Group
	if err = configs.Db.Where("group_number = ?", groupNumber).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "群聊不存在"})
		return
	}
	var member mychat.GroupMember
	err = configs.Db.Where("group_id = ? AND user_id = ?", group.ID, userID).First(&member).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "你不是该群成员"})
		return
	}
	// 群主不能直接退出（需解散或转让）
	if member.Role == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "群主不能退出群聊，请先转让群主或解散群聊"})
		return
	}
	// 删除成员记录
	if err = configs.Db.Delete(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "退出失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已退出群聊"})
}

// DismissGroup 解散群聊（仅群主）
func DismissGroup(c *gin.Context) {
	userID := c.GetUint("userID")
	groupNumber, err := strconv.Atoi(c.Param("group_number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的群号"})
		return
	}
	// 查询群组并验证群主身份
	var group mychat.Group
	if err = configs.Db.Where("group_number = ?", groupNumber).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "群聊不存在"})
		return
	}
	if group.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有群主可以解散群聊"})
		return
	}
	// 删除所有群成员（软删除或硬删除）
	if err = configs.Db.Where("group_id = ?", group.ID).Delete(&mychat.GroupMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解散失败"})
		return
	}
	// 删除群组（软删除）
	if err = configs.Db.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解散失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "群聊已解散"})
}

// GetGroupMembers 获取群成员列表（返回用户ID列表）
func GetGroupMembers(c *gin.Context) {
	groupNumber, err := strconv.Atoi(c.Param("group_number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的群号"})
		return
	}
	var group mychat.Group
	if err = configs.Db.Where("group_number = ?", groupNumber).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "群聊不存在"})
		return
	}
	var members []mychat.GroupMember
	if err = configs.Db.Where("group_id = ?", group.ID).Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		configs.Logger.Error("查询失败", zap.Any("err", err))
		return
	}

	// 收集用户ID
	userIDs := make([]uint, len(members))
	for i, m := range members {
		userIDs[i] = m.UserID
	}

	// 查询用户基本信息
	var users []model.User
	configs.Db.Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[uint]model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	type MemberInfo struct {
		UserID   uint   `json:"user_id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
		Role     string `json:"role"`
		JoinedAt string `json:"joined_at"`
	}
	result := make([]MemberInfo, 0)
	for _, m := range members {
		u := userMap[m.UserID]
		result = append(result, MemberInfo{
			UserID:   m.UserID,
			Username: u.Username,
			Avatar:   u.AvatarURL,
			Role:     m.Role,
			JoinedAt: m.JoinedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, result)
}

// GetMyGroups 获取用户加入的群聊列表
func GetMyGroups(c *gin.Context) {
	userID := c.GetUint("userID")

	var members []mychat.GroupMember
	if err := configs.Db.Where("user_id = ?", userID).Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		configs.Logger.Error("查询失败", zap.Any("err", err))
		return
	}

	groupIDs := make([]uint, len(members))
	for i, m := range members {
		groupIDs[i] = m.GroupID
	}

	var groups []mychat.Group
	configs.Db.Where("id IN ?", groupIDs).Find(&groups)

	c.JSON(http.StatusOK, groups)
}
