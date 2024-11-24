package handlers

import (
	"dbpilot/logger"
	"dbpilot/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// LogoutHandler はユーザーのログアウト処理を行うハンドラー
func LogoutHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}

		// リクエストのバリデーション
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Logger.Warn("不正なログアウトリクエスト",
				zap.Error(err),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// セッションの削除
		if err := models.DeleteSessionByEmail(db, req.Email); err != nil {
			logger.Logger.Error("セッション削除に失敗しました",
				zap.Error(err),
				zap.String("email", req.Email),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
			return
		}

		// 成功ログ
		logger.Logger.Info("ログアウト成功",
			zap.String("email", req.Email),
			zap.String("client_ip", c.ClientIP()),
		)

		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
	}
}
