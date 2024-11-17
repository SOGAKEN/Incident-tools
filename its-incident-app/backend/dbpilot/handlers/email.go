package handlers

import (
	"net/http"

	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddEmailHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// リクエストデータを受け取るための一時構造体
		var payload struct {
			MessageID string           `json:"message_id"`
			EmailData models.EmailData `json:"email_data"`
		}

		// JSONをバインド
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Payloadのmessage_idをEmailDataにセット
		emailData := payload.EmailData
		emailData.MessageID = payload.MessageID

		// データベースに保存
		if err := db.Create(&emailData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save email data"})
			return
		}

		// 保存成功時のレスポンス
		c.JSON(http.StatusOK, gin.H{
			"message": "Email data saved successfully",
			"data":    emailData,
		})
	}
}
