package handlers

import (
	"net/http"

	"dbpilot/logger"
	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func AddEmailHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			MessageID string           `json:"message_id"`
			EmailData models.EmailData `json:"email_data"`
		}

		// 共通のログフィールドを設定
		logFields := []zap.Field{
			zap.String("handler", "AddEmailHandler"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		// JSONをバインド
		if err := c.ShouldBindJSON(&payload); err != nil {
			logger.Logger.Error("リクエストのバインドに失敗しました",
				append(logFields,
					zap.Error(err))...)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// MessageIDをログフィールドに追加
		logFields = append(logFields, zap.String("message_id", payload.MessageID))
		logger.Logger.Info("メールデータの保存を開始します", logFields...)

		// Payloadのmessage_idをEmailDataにセット
		emailData := payload.EmailData
		emailData.MessageID = payload.MessageID

		// データベースに保存
		if err := db.Create(&emailData).Error; err != nil {
			logger.Logger.Error("メールデータの保存に失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save email data"})
			return
		}

		logger.Logger.Info("メールデータを保存しました",
			append(logFields,
				zap.Int("email_id", int(emailData.ID)),
				zap.String("subject", emailData.Subject))...)

		// 保存成功時のレスポンス
		c.JSON(http.StatusOK, gin.H{
			"message": "Email data saved successfully",
			"data":    emailData,
		})
	}
}
