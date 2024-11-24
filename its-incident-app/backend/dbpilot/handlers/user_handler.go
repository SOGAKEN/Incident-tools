package handlers

import (
	"net/http"

	"dbpilot/logger"
	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

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

// SaveUser はユーザー情報をDBに保存するハンドラー
func SaveUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Logger.Warn("不正なユーザー作成リクエスト",
				zap.Error(err),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// メールアドレスの重複チェック
		var existingUser models.User
		if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			logger.Logger.Warn("メールアドレスが既に使用されています",
				zap.String("email", req.Email),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		} else if err != gorm.ErrRecordNotFound {
			logger.Logger.Error("ユーザー重複チェックでエラーが発生",
				zap.Error(err),
				zap.String("email", req.Email),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
			return
		}

		// ユーザー作成
		user := models.User{
			Email:    req.Email,
			Password: req.Password, // 注: パスワードは既にハッシュ化されていることを前提
		}

		if err := db.Create(&user).Error; err != nil {
			logger.Logger.Error("ユーザー作成に失敗",
				zap.Error(err),
				zap.String("email", req.Email),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		logger.Logger.Info("新規ユーザーを作成しました",
			zap.Uint("user_id", user.ID),
			zap.String("email", user.Email),
		)

		c.JSON(http.StatusOK, gin.H{
			"message": "User saved successfully",
			"user_id": user.ID,
		})
	}
}

// QueryUser はユーザー情報を検索するハンドラー
func QueryUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req QueryUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Logger.Warn("不正なユーザー検索リクエスト",
				zap.Error(err),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		logger.Logger.Info("ユーザー検索リクエストを受信",
			zap.String("email", req.Email),
		)

		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				logger.Logger.Warn("ユーザーが見つかりません",
					zap.String("email", req.Email),
				)
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			logger.Logger.Error("ユーザー検索でエラーが発生",
				zap.Error(err),
				zap.String("email", req.Email),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query user"})
			return
		}

		logger.Logger.Info("ユーザー情報を取得しました",
			zap.Uint("user_id", user.ID),
			zap.String("email", user.Email),
		)

		c.JSON(http.StatusOK, QueryUserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Password: user.Password,
		})
	}
}
