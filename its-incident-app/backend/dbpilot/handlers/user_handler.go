package handlers

import (
	"net/http"

	"dbpilot/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ユーザー情報をDBに保存するためのハンドラー
type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type QueryUserRequest struct {
	Email string `json:"email"`
}

type QueryUserResponse struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SaveUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// すでにハッシュ化されたパスワードをそのまま保存
		user := models.User{Email: req.Email, Password: req.Password}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User saved successfully"})
	}
}
func QueryUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req QueryUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		var user models.User
		// Emailに基づいてユーザーを検索
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// ユーザー情報をレスポンスとして返す（パスワードはハッシュ化されたもの）
		c.JSON(http.StatusOK, QueryUserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Password: user.Password,
		})
	}
}
