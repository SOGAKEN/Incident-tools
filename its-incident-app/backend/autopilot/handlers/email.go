package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"autopilot/logger"
	"autopilot/models"
	"autopilot/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EmailHandler struct {
	dbpilotService *services.DBPilotService
	aiService      *services.AIService
}

func NewEmailHandler(dbpilot *services.DBPilotService, ai *services.AIService) *EmailHandler {
	return &EmailHandler{
		dbpilotService: dbpilot,
		aiService:      ai,
	}
}

func (h *EmailHandler) HandleEmailReceive(c *gin.Context) {
	messageID := c.GetHeader("X-Message-ID")
	if messageID == "" {
		logger.Logger.Warn("X-Message-IDヘッダーが存在しません")
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Message-ID header is required"})
		return
	}

	// 共通のログフィールドを設定
	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("handler", "HandleEmailReceive"),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	}

	var emailData models.EmailData
	if err := c.ShouldBindJSON(&emailData); err != nil {
		logger.Logger.Error("リクエストのバインドに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 処理状態の初期化
	status := models.NewProcessingStatus(messageID)
	if err := h.dbpilotService.UpdateProcessingStatus(status); err != nil {
		logger.Logger.Error("処理状態の初期化に失敗しました",
			append(logFields, zap.Error(err))...)
		// 処理は継続する
	}

	// メールデータの保存
	if err := h.dbpilotService.SaveEmail(&emailData, messageID); err != nil {
		logger.Logger.Error("メールデータの保存に失敗しました",
			append(logFields, zap.Error(err))...)
		status.SetFailed(err)
		_ = h.dbpilotService.UpdateProcessingStatus(status)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save email data",
			"details": err.Error(),
		})
		return
	}

	logger.Logger.Info("メールデータを保存しました", logFields...)

	// 非同期処理を開始する前に202レスポンスを返す
	c.JSON(http.StatusAccepted, gin.H{
		"status":     "processing",
		"message":    "Email received and being processed",
		"message_id": messageID,
	})

	// AI処理を非同期で実行
	go h.processEmailAsync(messageID, &emailData, logFields)
}

func (h *EmailHandler) processEmailAsync(messageID string, emailData *models.EmailData, logFields []zap.Field) {
	processCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	logger.Logger.Info("非同期AI処理を開始します", logFields...)

	if err := h.processAIAndSaveIncident(processCtx, emailData, messageID); err != nil {
		logger.Logger.Error("AI処理とインシデント保存に失敗しました",
			append(logFields, zap.Error(err))...)

		// エラー状態を保存
		status := &models.ProcessingStatus{
			MessageID: messageID,
		}
		status.SetFailed(err)
		if updateErr := h.dbpilotService.UpdateProcessingStatus(status); updateErr != nil {
			logger.Logger.Error("エラー状態の更新に失敗しました",
				append(logFields, zap.Error(updateErr))...)
		}
		return
	}

	// 処理完了を保存
	status := &models.ProcessingStatus{
		MessageID: messageID,
	}
	status.SetComplete()
	if err := h.dbpilotService.UpdateProcessingStatus(status); err != nil {
		logger.Logger.Error("完了状態の更新に失敗しました",
			append(logFields, zap.Error(err))...)
	}

	logger.Logger.Info("非同期AI処理が完了しました", logFields...)
}

func (h *EmailHandler) processAIAndSaveIncident(ctx context.Context, emailData *models.EmailData, messageID string) error {
	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("process", "AI_processing"),
	}

	// 処理状態を実行中に更新
	status := &models.ProcessingStatus{
		MessageID: messageID,
	}
	status.SetRunning("")
	if err := h.dbpilotService.UpdateProcessingStatus(status); err != nil {
		logger.Logger.Error("実行中状態の更新に失敗しました",
			append(logFields, zap.Error(err))...)
		// 処理は継続
	}

	logger.Logger.Info("AI処理を開始します", logFields...)

	// AI処理の実行
	aiResponse, err := h.aiService.ProcessEmail(ctx, emailData)
	if err != nil {
		logger.Logger.Error("AI処理に失敗しました",
			append(logFields, zap.Error(err))...)
		return err
	}

	// TaskIDを更新
	status.SetRunning(aiResponse.TaskID)
	if err := h.dbpilotService.UpdateProcessingStatus(status); err != nil {
		logger.Logger.Error("TaskIDの更新に失敗しました",
			append(logFields, zap.Error(err))...)
		// 処理は継続
	}

	logger.Logger.Info("AI処理が完了しました",
		append(logFields, zap.String("ai_task_id", aiResponse.TaskID))...)

	// インシデントの保存
	if err := h.dbpilotService.SaveIncident(aiResponse, messageID); err != nil {
		logger.Logger.Error("インシデントの保存に失敗しました",
			append(logFields,
				zap.String("ai_task_id", aiResponse.TaskID),
				zap.Error(err))...)
		return err
	}

	logger.Logger.Info("インシデントを保存しました",
		append(logFields, zap.String("ai_task_id", aiResponse.TaskID))...)
	return nil
}

func (h *EmailHandler) HandleCheckStatus(c *gin.Context) {
	messageID := c.Param("messageID")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message_id is required"})
		return
	}

	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("handler", "HandleCheckStatus"),
	}

	status, err := h.dbpilotService.GetProcessingStatus(messageID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Logger.Info("処理状態が見つかりません", logFields...)
			c.JSON(http.StatusNotFound, gin.H{
				"error":      "Processing status not found",
				"message_id": messageID,
			})
			return
		}

		logger.Logger.Error("処理状態の取得に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to get processing status",
			"message_id": messageID,
		})
		return
	}

	c.JSON(http.StatusOK, status)
}
