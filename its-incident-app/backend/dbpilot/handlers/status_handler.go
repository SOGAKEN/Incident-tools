package handlers

import (
	"net/http"
	"time"

	"dbpilot/logger"
	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UpdateProcessingStatus は処理状態を更新するハンドラー
func UpdateProcessingStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID := c.Param("messageID")
		if messageID == "" {
			logger.Logger.Warn("メッセージIDが指定されていません",
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "message_id is required"})
			return
		}

		var status models.ProcessingStatus
		if err := c.ShouldBindJSON(&status); err != nil {
			logger.Logger.Warn("不正なステータス更新リクエスト",
				zap.Error(err),
				zap.String("message_id", messageID),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Logger.Info("ステータス更新リクエストを受信",
			zap.String("message_id", messageID),
			zap.String("status", string(status.Status)),
			zap.String("task_id", status.TaskID),
		)

		// メッセージIDを上書き（URLパラメータを優先）
		status.MessageID = messageID

		// 既存のステータスを確認
		var existingStatus models.ProcessingStatus
		result := db.Where("message_id = ?", messageID).First(&existingStatus)

		if result.Error == gorm.ErrRecordNotFound {
			// 新規作成
			if err := db.Create(&status).Error; err != nil {
				logger.Logger.Error("ステータス作成に失敗",
					zap.Error(err),
					zap.String("message_id", messageID),
					zap.Any("status", status),
				)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			logger.Logger.Info("新規ステータスを作成しました",
				zap.String("message_id", messageID),
				zap.String("status", string(status.Status)),
				zap.String("task_id", status.TaskID),
			)
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
				logger.Logger.Info("処理が完了しました",
					zap.String("message_id", messageID),
					zap.String("status", string(status.Status)),
					zap.Time("completed_at", now),
				)
			}

			if err := db.Model(&existingStatus).Updates(updates).Error; err != nil {
				logger.Logger.Error("ステータス更新に失敗",
					zap.Error(err),
					zap.String("message_id", messageID),
					zap.Any("updates", updates),
				)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			logger.Logger.Info("ステータスを更新しました",
				zap.String("message_id", messageID),
				zap.String("status", string(status.Status)),
				zap.String("task_id", status.TaskID),
			)
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
			logger.Logger.Warn("メッセージIDが指定されていません",
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "message_id is required"})
			return
		}

		logger.Logger.Info("ステータス取得リクエストを受信",
			zap.String("message_id", messageID),
		)

		var status models.ProcessingStatus
		if err := db.Where("message_id = ?", messageID).First(&status).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				logger.Logger.Warn("ステータスが見つかりません",
					zap.String("message_id", messageID),
				)
				c.JSON(http.StatusNotFound, gin.H{"error": "Processing status not found"})
				return
			}
			logger.Logger.Error("ステータス取得に失敗",
				zap.Error(err),
				zap.String("message_id", messageID),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Logger.Info("ステータスを取得しました",
			zap.String("message_id", messageID),
			zap.String("status", string(status.Status)),
			zap.String("task_id", status.TaskID),
		)

		c.JSON(http.StatusOK, status)
	}
}
