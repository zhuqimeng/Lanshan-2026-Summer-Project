package info

import (
	"LanshanSummerProject/app/api/configs"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	model "LanshanSummerProject/app/api/internal/model/user"

	"github.com/gin-gonic/gin"
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

const (
	maxFileSize   = 10 << 20                 // 10MB
	uploadDir     = "Storage/picture/avatar" // 存储目录
	avatarBaseURL = "/avatar"                // 访问 URL 前缀
)

func UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size too big"})
		return
	}
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()
	buffer := make([]byte, 512)
	if _, err := src.Read(buffer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	mineType := http.DetectContentType(buffer)
	if !allowedMimeTypes[mineType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported Media Type"})
		return
	}
	src.Seek(0, 0)
	ext := strings.ToLower(path.Ext(file.Filename))
	ext = getExtByMime(mineType)
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	savePath := path.Join(uploadDir, filename)
	relativePath := avatarBaseURL + "/" + filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user := &model.User{}
	if err := configs.Db.First(user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user.AvatarURL != "" && user.AvatarURL != "/avatar/default.png" {
		oldPath := strings.TrimPrefix(user.AvatarURL, avatarBaseURL)
		oldFullPath := filepath.Join(uploadDir, oldPath)
		_ = os.Remove(oldFullPath)
	}
	user.AvatarURL = relativePath
	if err := configs.Db.Save(user).Error; err != nil {
		_ = os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"avatar_url": relativePath,
		"message":    "上传成功",
	})
}

func getExtByMime(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}
