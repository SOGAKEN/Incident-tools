package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"autopilot/logger"
	"autopilot/models"
	"autopilot/services"
	"autopilot/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EmailHandler struct {
	dbpilotService *services.DBPilotService
	aiService      *services.AIService
	emailStore     *store.EmailStore
}

func NewEmailHandler(dbpilot *services.DBPilotService, ai *services.AIService, projectID string) *EmailHandler {
	emailStore, err := store.NewEmailStore(context.Background(), projectID)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize email store", zap.Error(err))
	}

	return &EmailHandler{
		dbpilotService: dbpilot,
		aiService:      ai,
		emailStore:     emailStore,
	}
}

func (h *EmailHandler) HandleEmailReceive(c *gin.Context) {
	ctx := c.Request.Context()
	messageID := c.GetHeader("X-Message-ID")
	if messageID == "" {
		logger.Logger.Warn("X-Message-IDヘッダーが存在しません")
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Message-ID header is required"})
		return
	}

	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("handler", "HandleEmailReceive"),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	}

	// Datastoreに初期状態を作成
	if err := h.emailStore.CreateProcessing(ctx, messageID); err != nil {
		logger.Logger.Error("処理状態の初期化に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize processing"})
		return
	}

	var emailData models.EmailData
	if err := c.ShouldBindJSON(&emailData); err != nil {
		logger.Logger.Error("リクエストのバインドに失敗しました",
			append(logFields, zap.Error(err))...)
		h.emailStore.SetError(ctx, messageID, "BIND_ERROR", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// メールデータを保存
	if err := h.dbpilotService.SaveEmail(&emailData, messageID); err != nil {
		logger.Logger.Error("メールデータの保存に失敗しました",
			append(logFields, zap.Error(err))...)
		h.emailStore.SetError(ctx, messageID, "SAVE_ERROR", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save email data",
			"details": err.Error(),
		})
		return
	}

	logger.Logger.Debug("メールデータを保存しました", logFields...)

	// 202レスポンスを返す
	c.JSON(http.StatusAccepted, gin.H{
		"status":     "processing",
		"message":    "Email received and being processed",
		"message_id": messageID,
	})

	// 非同期処理を開始
	go h.processEmailAsync(messageID, &emailData, logFields)
}

func (h *EmailHandler) processEmailAsync(messageID string, emailData *models.EmailData, logFields []zap.Field) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	logger.Logger.Debug("非同期AI処理を開始します", logFields...)

	if err := h.processAIAndSaveIncident(ctx, emailData, messageID); err != nil {
		logger.Logger.Error("AI処理とインシデント保存に失敗しました",
			append(logFields, zap.Error(err))...)
		h.emailStore.SetError(ctx, messageID, "AI_PROCESS_ERROR", err.Error())
		return
	}

	// 完了状態を設定
	processing, err := h.emailStore.GetProcessing(ctx, messageID)
	if err == nil && processing != nil {
		processing.SetComplete()
		h.emailStore.UpdateProcessing(ctx, processing)
	}

	logger.Logger.Debug("非同期AI処理が完了しました", logFields...)
}

func (h *EmailHandler) processAIAndSaveIncident(ctx context.Context, emailData *models.EmailData, messageID string) error {
	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("process", "AI_processing"),
	}

	// 実行中状態に更新
	processing, _ := h.emailStore.GetProcessing(ctx, messageID)
	serviceState, _ := h.emailStore.GetServiceState(ctx, messageID)

	if processing != nil {
		processing.SetRunning()
		h.emailStore.UpdateProcessing(ctx, processing)
	}

	if serviceState != nil {
		serviceState.SetRunning("")
		h.emailStore.UpdateServiceState(ctx, serviceState)
	}

	logger.Logger.Info("AI処理を開始します", logFields...)

	// AI処理の実行
	aiResponse, err := h.aiService.ProcessEmail(ctx, emailData)
	if err != nil {
		logger.Logger.Error("AI処理に失敗しました",
			append(logFields, zap.Error(err))...)

		errorResponse := models.NewErrorResponse(messageID, err)
		if saveErr := h.dbpilotService.SaveIncident(errorResponse, messageID); saveErr != nil {
			logger.Logger.Error("エラー情報のインシデント保存に失敗しました",
				append(logFields,
					zap.Error(saveErr),
					zap.Error(err))...)
			return fmt.Errorf("failed to save error incident: %v (original error: %v)", saveErr, err)
		}

		return err
	}

	logger.Logger.Debug("AI処理のレスポンス",
		append(logFields, zap.Any("ai_response", aiResponse))...)

	// サービス状態を更新
	if serviceState != nil {
		serviceState.SetRunning(aiResponse.TaskID)
		h.emailStore.UpdateServiceState(ctx, serviceState)
	}

	logger.Logger.Info("AI処理が完了しました",
		append(logFields, zap.String("task_id", aiResponse.TaskID))...)

	// インシデントを保存
	if err := h.dbpilotService.SaveIncident(aiResponse, messageID); err != nil {
		logger.Logger.Error("インシデントの保存に失敗しました",
			append(logFields,
				zap.String("task_id", aiResponse.TaskID),
				zap.Error(err))...)
		return err
	}

	// AI処理の結果を保存
	if serviceState != nil {
		serviceState.SetComplete(aiResponse)
		h.emailStore.UpdateServiceState(ctx, serviceState)
	}

	logger.Logger.Debug("インシデントを保存しました",
		append(logFields, zap.String("task_id", aiResponse.TaskID))...)
	return nil
}

func (h *EmailHandler) HandleCheckStatus(c *gin.Context) {
	ctx := c.Request.Context()
	messageID := c.Param("messageID")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message_id is required"})
		return
	}

	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("handler", "HandleCheckStatus"),
	}

	// 全体の処理状態を取得
	processing, err := h.emailStore.GetProcessing(ctx, messageID)
	if err != nil {
		logger.Logger.Error("処理状態の取得に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get processing status"})
		return
	}

	// サービスの状態を取得
	serviceState, err := h.emailStore.GetServiceState(ctx, messageID)
	if err != nil {
		logger.Logger.Error("サービス状態の取得に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get service state"})
		return
	}

	if processing == nil || serviceState == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message_id": messageID,
		"processing": processing,
		"service":    serviceState,
	})
}
