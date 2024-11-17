package handlers

import (
	"net/http"

	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProfileRequest はプロフィールの登録に使用されるリクエスト構造体
type ProfileRequest struct {
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

type ProfileResponse struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

// RegisterProfile はセッションからUserIDを取得し、プロフィールを登録します
func RegisterProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, exists := c.Get("session")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Session not found"})
			return
		}

		// セッションIDからUserIDを取得
		var session models.LoginSession
		if err := db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			return
		}

		// リクエストのバリデーション
		var req ProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// プロフィールの登録
		profile := models.Profile{UserID: session.UserID, Name: req.Name, ImageURL: req.ImageURL}
		if err := db.Create(&profile).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile created successfully"})
	}
}

// GetProfile はセッションIDを使ってユーザーのプロフィール情報を取得します
func GetProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, exists := c.Get("session")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Session not found"})
			return
		}

		var session models.LoginSession
		if err := db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			return
		}

		// ユーザーとそのプロフィール情報を取得
		var user models.User
		if err := db.Preload("Profile").Where("id = ?", session.UserID).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User or profile not found"})
			return
		}

		// プロフィール情報をレスポンス
		c.JSON(http.StatusOK, ProfileResponse{
			UserID:   user.ID,
			Email:    user.Email,
			Name:     user.Profile.Name,
			ImageURL: user.Profile.ImageURL,
		})
	}
}
