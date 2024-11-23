package handlers

import (
	"net/http"

	"dbpilot/logger"
	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
		// セッション情報の取得
		sessionID, exists := c.Get("session")
		if !exists {
			logger.Logger.Error("セッション情報が見つかりません")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Session not found"})
			return
		}

		// 文字列型へ変換
		sessionIDStr, ok := sessionID.(string)
		if !ok {
			logger.Logger.Error("セッションIDの型が不正です")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid session format"})
			return
		}

		// セッションの検証
		var session models.LoginSession
		if err := db.Where("session_id = ?", sessionIDStr).First(&session).Error; err != nil {
			logger.Logger.Error("セッションの検証に失敗しました",
				zap.Error(err),
				zap.String("session_id", sessionIDStr),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			return
		}

		// ユーザーとプロフィール情報の取得
		var user models.User
		if err := db.Preload("Profile").Where("id = ?", session.UserID).First(&user).Error; err != nil {
			logger.Logger.Error("ユーザーまたはプロフィール情報の取得に失敗しました",
				zap.Error(err),
				zap.Uint("user_id", session.UserID),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": "User or profile not found"})
			return
		}

		c.JSON(http.StatusOK, ProfileResponse{
			UserID:   user.ID,
			Email:    user.Email,
			Name:     user.Profile.Name,
			ImageURL: user.Profile.ImageURL,
		})
	}
}
