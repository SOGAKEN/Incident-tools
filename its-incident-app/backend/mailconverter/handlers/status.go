package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mailconvertor/logger"
	"mailconvertor/store"
)

func HandleCheckStatus(c *gin.Context) {
	log := logger.Logger
	ctx := c.Request.Context()
	messageID := c.Param("messageID")

	// EmailStoreの初期化
	emailStore, err := store.NewEmailStore(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Error("Failed to initialize email store", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}
	defer emailStore.Close()

	// 全体の処理状態を取得
	processing, err := emailStore.GetProcessing(ctx, messageID)
	if err != nil {
		log.Error("Failed to get processing state",
			zap.String("messageId", messageID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get processing state",
		})
		return
	}

	// サービスの状態を取得
	serviceState, err := emailStore.GetServiceState(ctx, messageID)
	if err != nil {
		log.Error("Failed to get service state",
			zap.String("messageId", messageID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get service state",
		})
		return
	}

	if processing == nil || serviceState == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Message not found",
		})
		return
	}

	// レスポンスの構築
	response := gin.H{
		"message_id": messageID,
		"processing": gin.H{
			"status":     processing.Status,
			"created_at": processing.CreatedAt,
			"updated_at": processing.UpdatedAt,
		},
		"service_state": gin.H{
			"service_type":  serviceState.ServiceType,
			"status":        serviceState.Status,
			"error_code":    serviceState.ErrorCode,
			"error_message": serviceState.ErrorMessage,
			"created_at":    serviceState.CreatedAt,
			"updated_at":    serviceState.UpdatedAt,
		},
	}

	// メールデータが存在する場合は追加
	if serviceState.EmailData != nil {
		response["email_data"] = gin.H{
			"from":    serviceState.EmailData.From,
			"to":      serviceState.EmailData.To,
			"subject": serviceState.EmailData.Subject,
			"date":    serviceState.EmailData.Date,
		}
	}

	c.JSON(http.StatusOK, response)
}
