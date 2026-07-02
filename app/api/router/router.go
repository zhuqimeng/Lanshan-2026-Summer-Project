package router

import (
	"LanshanSummerProject/app/api/configs"
	"LanshanSummerProject/app/api/internal/middleware"
	"LanshanSummerProject/app/api/internal/service/user/group"
	"LanshanSummerProject/app/api/internal/service/user/info"
	"LanshanSummerProject/app/api/internal/service/ws"
	"LanshanSummerProject/utils/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Start() {
	hub := ws.NewHub()
	go hub.Run()
	ws.InitWebSocket(hub)
	r := gin.Default()
	r.POST("/register", info.Register)
	r.POST("/login", info.Login)
	r.POST("/refresh", jwt.RefreshToken)
	auth := r.Group("/api")
	auth.Use(middleware.Authorize())
	{
		auth.DELETE("logout", info.Logout)
		auth.POST("/avatar/upload", info.UploadAvatar)
		auth.POST("/friends/requests", group.SendFriendRequest)
		auth.GET("/friends/requests", group.GetFriendRequests)
		auth.PUT("/friends/requests/:id", group.HandleFriendRequest) // 处理请求（同意/拒绝）
		auth.DELETE("/friends/:friend_id", group.DeleteFriend)
		auth.GET("/friends", group.GetFriendList)                        // 获取好友列表（支持?group=xxx）
		auth.PUT("/friends/:friend_id/remark", group.UpdateFriendRemark) // 修改备注
		auth.PUT("/friends/:friend_id/group", group.UpdateFriendGroup)   // 修改分组
		auth.GET("/friends/groups", group.GetFriendGroups)               // 获取所有分组名
		auth.GET("/chat", ws.WebSocketHandler)
		auth.GET("/messages/offline", ws.GetOfflineMessages)
		auth.PUT("/messages/read", ws.MarkMessagesAsRead)
		// 群聊管理（全部使用 group_number 作为路径参数）
		auth.POST("/groups", ws.CreateGroup)
		auth.POST("/groups/join", ws.JoinGroup)
		auth.GET("/groups/:group_number/messages", ws.GetGroupMessages)
		auth.DELETE("groups/:group_number/leave", ws.LeaveGroup)
		auth.DELETE("groups/:group_number/dismiss", ws.DismissGroup)
		auth.GET("/groups/:group_number/members", ws.GetGroupMembers)
		auth.GET("/groups/my", ws.GetMyGroups)
	}
	if err := r.Run(":8080"); err != nil {
		configs.Logger.Fatal("Gin run error", zap.Error(err))
	}
}
