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
		logger.Logger.Fatal("メールストアの初期化に失敗しました", zap.Error(err))
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
		logger.Logger.Info("メッセージIDが未指定です") // WARNからINFOに変更（想定内のエラー）
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Message-ID header is required"})
		return
	}

	// 共通のログフィールド
	baseFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("path", c.Request.URL.Path),
	}

	// Datastoreに初期状態を作成
	if err := h.emailStore.CreateProcessing(ctx, messageID); err != nil {
		logger.Logger.Error("処理状態の初期化に失敗しました",
			append(baseFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize processing"})
		return
	}

	// リクエストのバインド
	var emailData models.EmailData
	if err := c.ShouldBindJSON(&emailData); err != nil {
		logger.Logger.Info("不正なリクエスト形式です", // ERRORからINFOに変更（想定内のエラー）
			append(baseFields, zap.Error(err))...)
		h.emailStore.SetError(ctx, messageID, "BIND_ERROR", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// メールデータを保存
	if err := h.dbpilotService.SaveEmail(&emailData, messageID); err != nil {
		logger.Logger.Error("メールデータの保存に失敗しました",
			append(baseFields, zap.Error(err))...)
		h.emailStore.SetError(ctx, messageID, "SAVE_ERROR", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save email data"})
		return
	}

	// 非同期処理開始のログ
	logger.Logger.Info("メール処理を開始します", baseFields...)

	c.JSON(http.StatusAccepted, gin.H{
		"status":     "processing",
		"message_id": messageID,
	})

	go h.processEmailAsync(messageID, &emailData, baseFields)
}

func (h *EmailHandler) processEmailAsync(messageID string, emailData *models.EmailData, logFields []zap.Field) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if err := h.processAIAndSaveIncident(ctx, emailData, messageID); err != nil {
		logger.Logger.Error("AI処理が失敗しました",
			append(logFields, zap.Error(err))...)
		h.emailStore.SetError(ctx, messageID, "AI_PROCESS_ERROR", err.Error())
		return
	}

	// 処理完了時の状態更新
	if processing, err := h.emailStore.GetProcessing(ctx, messageID); err == nil && processing != nil {
		processing.SetComplete()
		h.emailStore.UpdateProcessing(ctx, processing)
		logger.Logger.Info("メール処理が完了しました", logFields...)
	}
}

func (h *EmailHandler) processAIAndSaveIncident(ctx context.Context, emailData *models.EmailData, messageID string) error {
	logFields := []zap.Field{
		zap.String("message_id", messageID),
	}

	// 処理状態の更新
	if processing, _ := h.emailStore.GetProcessing(ctx, messageID); processing != nil {
		processing.SetRunning()
		h.emailStore.UpdateProcessing(ctx, processing)
	}
	if serviceState, _ := h.emailStore.GetServiceState(ctx, messageID); serviceState != nil {
		serviceState.SetRunning("")
		h.emailStore.UpdateServiceState(ctx, serviceState)
	}

	// AI処理の実行 - messageIDを渡すように修正
	aiResponse, err := h.aiService.ProcessEmail(ctx, emailData, messageID)
	if err != nil {
		// エラーレスポンスの保存
		errorResponse := models.NewErrorResponse(messageID, err)
		if saveErr := h.dbpilotService.SaveIncident(errorResponse, messageID); saveErr != nil {
			logger.Logger.Error("エラー情報の保存に失敗しました",
				append(logFields,
					zap.Error(saveErr),
					zap.Error(err))...)
		}
		return err
	}

	// 処理状態の更新
	if serviceState, _ := h.emailStore.GetServiceState(ctx, messageID); serviceState != nil {
		serviceState.SetComplete(aiResponse)
		h.emailStore.UpdateServiceState(ctx, serviceState)
	}

	// インシデントの保存
	if err := h.dbpilotService.SaveIncident(aiResponse, messageID); err != nil {
		return fmt.Errorf("failed to save incident: %v", err)
	}

	return nil
}

func (h *EmailHandler) HandleCheckStatus(c *gin.Context) {
	ctx := c.Request.Context()
	messageID := c.Param("messageID")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message_id is required"})
		return
	}

	logFields := []zap.Field{zap.String("message_id", messageID)}

	// 処理状態の取得
	processing, err := h.emailStore.GetProcessing(ctx, messageID)
	if err != nil {
		logger.Logger.Error("処理状態の取得に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get processing status"})
		return
	}

	// サービス状態の取得
	serviceState, err := h.emailStore.GetServiceState(ctx, messageID)
	if err != nil {
		logger.Logger.Error("サービス状態の取得に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get service state"})
		return
	}

	if processing == nil || serviceState == nil {
		logger.Logger.Info("状態情報が見つかりません", logFields...)
		c.JSON(http.StatusNotFound, gin.H{"error": "Status not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message_id": messageID,
		"processing": processing,
		"service":    serviceState,
	})
}
