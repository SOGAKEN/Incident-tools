package handlers

import (
	"net/http"
	"time"

	"dbpilot/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UpdateProcessingStatus は処理状態を更新するハンドラー
func UpdateProcessingStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID := c.Param("messageID")
		if messageID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "message_id is required"})
			return
		}

		var status models.ProcessingStatus
		if err := c.ShouldBindJSON(&status); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// メッセージIDを上書き（URLパラメータを優先）
		status.MessageID = messageID

		// 既存のステータスを確認
		var existingStatus models.ProcessingStatus
		result := db.Where("message_id = ?", messageID).First(&existingStatus)

		if result.Error == gorm.ErrRecordNotFound {
			// 新規作成
			if err := db.Create(&status).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			// 更新
			updates := map[string]interface{}{
				"status":  status.Status,
				"task_id": status.TaskID,
				"error":   status.Error,
			}

			if status.Status == models.StatusComplete || status.Status == models.StatusFailed {
				now := time.Now()
				updates["completed_at"] = &now
			}

			if err := db.Model(&existingStatus).Updates(updates).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Processing status updated successfully",
			"status":  status,
		})
	}
}

// GetProcessingStatus は処理状態を取得するハンドラー
func GetProcessingStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID := c.Param("messageID")
		if messageID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "message_id is required"})
			return
		}

		var status models.ProcessingStatus
		if err := db.Where("message_id = ?", messageID).First(&status).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Processing status not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, status)
	}
}
