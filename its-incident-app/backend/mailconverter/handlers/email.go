package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"mailconvertor/logger"
	"mailconvertor/models"
	"mailconvertor/store"

	"github.com/gin-gonic/gin"
	"github.com/jhillyerd/enmime"
	"go.uber.org/zap"
)

type EmailHandler struct {
}

func NewEmailHandler() *EmailHandler {
	return &EmailHandler{}
}

func ParseEmail(rawEmailData []byte) (*models.EmailData, error) {
	log := logger.Logger

	reader := bytes.NewReader(rawEmailData)
	env, err := enmime.ReadEnvelope(reader)
	if err != nil {
		log.Error("MIMEメッセージのパースに失敗しました", zap.Error(err))
		return nil, fmt.Errorf("failed to parse MIME message: %v", err)
	}

	emailData := &models.EmailData{
		From:                    env.GetHeader("From"),
		To:                      env.GetHeader("To"),
		Subject:                 env.GetHeader("Subject"),
		Date:                    env.GetHeader("Date"),
		OriginalMessageID:       env.GetHeader("Message-ID"),
		MIMEVersion:             env.GetHeader("MIME-Version"),
		ContentType:             env.GetHeader("Content-Type"),
		ContentTransferEncoding: env.GetHeader("Content-Transfer-Encoding"),
		CC:                      env.GetHeader("CC"),
		Body:                    env.Text,
	}

	if len(env.Attachments) > 0 {
		emailData.FileName = env.Attachments[0].FileName
	}

	log.Debug("メールのパースが完了しました",
		zap.String("messageId", emailData.OriginalMessageID),
		zap.String("from", emailData.From),
		zap.String("subject", emailData.Subject),
	)

	return emailData, nil
}

func createResponse(status string, code int, message string, traceID string, err error) models.APIResponse {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	response := models.APIResponse{
		Status:    status,
		Code:      code,
		Message:   message,
		TraceID:   traceID,
		Timestamp: timestamp,
	}

	if err != nil {
		errType := "unknown_error"
		switch {
		case code == http.StatusBadRequest:
			errType = "invalid_request"
		case code == http.StatusInternalServerError:
			errType = "internal_error"
		}

		response.Error = &models.ErrorInfo{
			Type:    errType,
			Message: err.Error(),
			Detail:  fmt.Sprintf("%+v", err),
		}
	}

	return response
}

func (h *EmailHandler) HandleEmailReceive(c *gin.Context) {
	log := logger.Logger
	ctx := c.Request.Context()

	messageID := c.GetHeader("X-Message-ID")
	if messageID == "" {
		messageID = fmt.Sprintf("gen-%d", time.Now().UnixNano())
		log.Info("メッセージIDを生成しました", zap.String("messageId", messageID))
	}

	log.Info("メール受信処理を開始します", zap.String("messageId", messageID))

	// EmailStoreの初期化
	emailStore, err := store.NewEmailStore(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Error("Failed to initialize email store", zap.Error(err))
		response := createResponse("error", http.StatusInternalServerError, "Internal server error", messageID, err)
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer emailStore.Close()

	// 初期状態の作成
	if err := emailStore.CreateProcessing(ctx, messageID); err != nil {
		log.Error("Failed to create initial processing", zap.Error(err))
		response := createResponse("error", http.StatusInternalServerError, "Internal server error", messageID, err)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// リクエスト情報のログ出力
	log.Info("リクエスト情報",
		zap.String("messageId", messageID),
		zap.String("method", c.Request.Method),
		zap.String("content-type", c.GetHeader("Content-Type")),
		zap.Int64("content-length", c.Request.ContentLength),
	)

	// パニックハンドラー
	defer func() {
		if r := recover(); r != nil {
			log.Error("パニックが発生しました",
				zap.String("messageId", messageID),
				zap.Any("error", r),
				zap.String("stack", string(debug.Stack())))

			if err := emailStore.SetError(ctx, messageID, "PANIC", fmt.Sprintf("%v", r)); err != nil {
				log.Error("パニック状態の保存に失敗しました", zap.Error(err))
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
	}()

	// リクエストボディの状態確認
	if c.Request.Body == nil {
		err := fmt.Errorf("request body is nil")
		if err := emailStore.SetError(ctx, messageID, "EMPTY_BODY", err.Error()); err != nil {
			log.Error("エラー状態の保存に失敗しました", zap.Error(err))
		}
		response := createResponse("error", http.StatusBadRequest, "Request body is nil", messageID, err)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	rawEmailData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("リクエストボディの読み取りに失敗しました",
			zap.String("messageId", messageID),
			zap.Error(err))
		if err := emailStore.SetError(ctx, messageID, "READ_ERROR", err.Error()); err != nil {
			log.Error("エラー状態の保存に失敗しました", zap.Error(err))
		}
		response := createResponse("error", http.StatusBadRequest, "Failed to read request body", messageID, err)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if len(rawEmailData) == 0 {
		err := fmt.Errorf("request data is empty")
		if err := emailStore.SetError(ctx, messageID, "EMPTY_DATA", err.Error()); err != nil {
			log.Error("エラー状態の保存に失敗しました", zap.Error(err))
		}
		response := createResponse("error", http.StatusBadRequest, "Request data is empty", messageID, err)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	log.Info("メールデータを受信しました",
		zap.String("messageId", messageID),
		zap.Int("size", len(rawEmailData)))

	// メール処理状態を実行中に更新
	processing, err := emailStore.GetProcessing(ctx, messageID)
	if err != nil {
		log.Error("処理状態の取得に失敗しました", zap.Error(err))
	} else {
		processing.Status = models.StatusRunning
		if err := emailStore.UpdateProcessing(ctx, processing); err != nil {
			log.Error("処理状態の更新に失敗しました", zap.Error(err))
		}
	}

	// サービス状態も実行中に更新
	serviceState, err := emailStore.GetServiceState(ctx, messageID)
	if err != nil {
		log.Error("サービス状態の取得に失敗しました", zap.Error(err))
	} else {
		serviceState.Status = models.StatusRunning
		if err := emailStore.UpdateServiceState(ctx, serviceState); err != nil {
			log.Error("サービス状態の更新に失敗しました", zap.Error(err))
		}
	}

	emailData, err := ParseEmail(rawEmailData)
	if err != nil {
		log.Error("メールのパースに失敗しました",
			zap.String("messageId", messageID),
			zap.Error(err))
		if err := emailStore.SetError(ctx, messageID, "PARSE_ERROR", err.Error()); err != nil {
			log.Error("エラー状態の保存に失敗しました", zap.Error(err))
		}
		response := createResponse("error", http.StatusInternalServerError, "Failed to parse email", messageID, err)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	log.Info("メールのパースが完了しました",
		zap.String("messageId", messageID))

	// パース済みメールデータを状態に保存
	serviceState.EmailData = emailData
	if err := emailStore.UpdateServiceState(ctx, serviceState); err != nil {
		log.Error("メールデータの保存に失敗しました", zap.Error(err))
	}

	// 外部APIへの送信
	log.Info("AutoPilotへの送信を開始します",
		zap.String("messageId", messageID))

	if err := sendToExternalAPI(ctx, emailData, messageID); err != nil {
		log.Error("AutoPilotへの送信に失敗しました",
			zap.String("messageId", messageID),
			zap.Error(err))
		if err := emailStore.SetError(ctx, messageID, "API_ERROR", err.Error()); err != nil {
			log.Error("エラー状態の保存に失敗しました", zap.Error(err))
		}
		response := createResponse("error", http.StatusInternalServerError, "Failed to send to external API", messageID, err)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// 処理完了状態の保存
	processing.Status = models.StatusComplete
	if err := emailStore.UpdateProcessing(ctx, processing); err != nil {
		log.Error("完了状態の保存に失敗しました", zap.Error(err))
	}

	serviceState.Status = models.StatusComplete
	if err := emailStore.UpdateServiceState(ctx, serviceState); err != nil {
		log.Error("サービス完了状態の保存に失敗しました", zap.Error(err))
	}

	log.Info("メール処理が正常に完了しました",
		zap.String("messageId", messageID))
	response := createResponse("success", http.StatusOK, "Email processed successfully", messageID, nil)
	c.JSON(http.StatusOK, response)
}

func logEmailData(emailData *models.EmailData) {
	log := logger.Logger

	log.Debug("パースされたメールデータ",
		zap.String("messageId", emailData.OriginalMessageID),
		zap.String("from", emailData.From),
		zap.String("to", emailData.To),
		zap.String("subject", emailData.Subject),
		zap.String("date", emailData.Date),
		zap.String("contentType", emailData.ContentType),
		zap.Int("bodyLength", len(emailData.Body)),
		zap.Bool("hasFileName", emailData.FileName != ""),
	)
}

func sendToExternalAPI(ctx context.Context, emailData *models.EmailData, messageID string) error {
	log := logger.Logger

	payloadBytes, err := json.MarshalIndent(emailData, "", "  ")
	if err != nil {
		log.Error("ペイロードのJSONエンコードに失敗しました", zap.Error(err))
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	log.Info("AutoPilotにデータを送信します",
		zap.String("messageId", messageID),
		zap.String("originalMsgId", emailData.OriginalMessageID),
		zap.String("from", emailData.From),
		zap.String("to", emailData.To),
		zap.String("subject", emailData.Subject),
		zap.Int("payloadSize", len(payloadBytes)),
	)

	apiURL := os.Getenv("AUTOPILOT_URL")
	bearerToken := os.Getenv("SERVICE_TOKEN")
	if bearerToken == "" {
		log.Error("Bearer tokenが設定されていません")
		return fmt.Errorf("bearer token is not set")
	}

	req, err := http.NewRequest("POST", apiURL+"/receive", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Error("HTTPリクエストの作成に失敗しました", zap.Error(err))
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("X-Message-ID", messageID)

	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		log.Error("HTTPリクエストの実行に失敗しました", zap.Error(err))
		return fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	log.Info("AutoPilotからレスポンスを受信しました",
		zap.String("messageId", messageID),
		zap.Int("statusCode", resp.StatusCode),
		zap.String("status", resp.Status),
	)

	if resp.StatusCode >= 400 {
		log.Error("AutoPilotがエラーを返しました",
			zap.String("messageId", messageID),
			zap.Int("statusCode", resp.StatusCode))
		return fmt.Errorf("external API returned error status: %d", resp.StatusCode)
	}

	log.Info("AutoPilotにデータを送信しました",
		zap.String("messageId", messageID),
		zap.Int("statusCode", resp.StatusCode))
	return nil
}
