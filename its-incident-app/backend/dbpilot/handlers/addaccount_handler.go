// handlers/login_token_handler.go
package handlers

import (
	"dbpilot/models"
	"net/http"
	"time"

	"dbpilot/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ResponseWrapper struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// handleError はエラーレスポンスを統一的に処理
func handleError(c *gin.Context, statusCode int, err error, additionalFields ...zap.Field) {
	fields := append([]zap.Field{
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Error(err),
	}, additionalFields...)

	logger.Logger.Error("エラーが発生しました", fields...)

	c.JSON(statusCode, ResponseWrapper{
		Success: false,
		Error:   err.Error(),
	})
}

// handleSuccess は成功レスポンスを統一的に処理
func handleSuccess(c *gin.Context, data interface{}, additionalFields ...zap.Field) {
	fields := append([]zap.Field{
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	}, additionalFields...)

	logger.Logger.Info("処理が成功しました", fields...)

	c.JSON(200, ResponseWrapper{
		Success: true,
		Data:    data,
	})
}

func CreateLoginToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email     string    `json:"email" binding:"required,email"`
			Token     string    `json:"token" binding:"required"`
			ExpiresAt time.Time `json:"expires_at" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			handleError(c, http.StatusBadRequest, err)
			return
		}

		// トランザクションを開始
		err := db.Transaction(func(tx *gorm.DB) error {
			// ユーザーを検索または作成
			var user models.User
			if err := tx.FirstOrCreate(&user, models.User{
				Email: req.Email,
			}).Error; err != nil {
				return err
			}

			// 既存の未使用トークンを無効化
			if err := tx.Model(&models.LoginToken{}).
				Where("email = ? AND used = ? AND expires_at > ?",
					req.Email, false, time.Now()).
				Update("used", true).Error; err != nil {
				return err
			}

			// 新しいトークンを作成
			loginToken := &models.LoginToken{
				Email:     req.Email,
				Token:     req.Token,
				ExpiresAt: req.ExpiresAt,
				Used:      false,
			}

			if err := tx.Create(loginToken).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			handleError(c, http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login token created successfully",
			"email":   req.Email,
		})
	}
}

func VerifyLoginToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "VerifyLoginToken"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		token := c.Query("token")
		if token == "" {
			logger.Logger.Error("トークンが指定されていません", logFields...)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
			return
		}

		logFields = append(logFields, zap.String("token", token))

		// デバッグ用：トークンの状態を確認
		var loginToken models.LoginToken
		result := db.Where("token = ?", token).First(&loginToken)

		if result.Error != nil {
			logger.Logger.Error("トークンが見つかりません",
				append(logFields, zap.Error(result.Error))...)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// トークンの状態をログ出力
		logFields = append(logFields,
			zap.Time("expires_at", loginToken.ExpiresAt),
			zap.Bool("used", loginToken.Used),
			zap.String("email", loginToken.Email))

		logger.Logger.Info("トークンの状態", logFields...)

		// トークンの有効性チェック
		if loginToken.Used {
			logger.Logger.Error("トークンは既に使用済みです", logFields...)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has already been used"})
			return
		}

		if loginToken.ExpiresAt.Before(time.Now()) {
			logger.Logger.Error("トークンの有効期限が切れています", logFields...)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			return
		}

		// トークンを使用済みにマーク
		if err := db.Model(&loginToken).Update("used", true).Error; err != nil {
			logger.Logger.Error("トークンの更新に失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token status"})
			return
		}

		// ユーザー情報を取得
		var user models.User
		if err := db.Where("email = ?", loginToken.Email).First(&user).Error; err != nil {
			logger.Logger.Error("ユーザーが見つかりません",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		logger.Logger.Info("トークンの検証が成功しました",
			append(logFields,
				zap.Uint("user_id", user.ID))...)

		c.JSON(http.StatusOK, gin.H{
			"message": "Token verified successfully",
			"email":   user.Email,
			"user_id": user.ID,
		})
	}
}
