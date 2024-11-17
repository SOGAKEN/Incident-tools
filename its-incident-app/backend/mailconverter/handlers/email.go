package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jhillyerd/enmime"
	"go.uber.org/zap"
	"mailconvertor/logger"
	"mailconvertor/models"
)

func ParseEmail(rawEmailData []byte) (*models.EmailData, error) {
	// ロガーの取得
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

func HandleEmailReceive(c *gin.Context) {
	// ロガーの取得
	log := logger.Logger

	messageID := c.GetHeader("X-Message-ID")
	if messageID == "" {
		messageID = fmt.Sprintf("gen-%d", time.Now().UnixNano())
		log.Info("メッセージIDを生成しました", zap.String("messageId", messageID))
	}

	rawEmailData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("リクエストボディの読み取りに失敗しました", zap.Error(err))
		response := createResponse("error", http.StatusBadRequest, "Failed to read request body", messageID, err)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	log.Debug("メールデータを受信しました",
		zap.String("messageId", messageID),
		zap.Int("size", len(rawEmailData)),
	)

	emailData, err := ParseEmail(rawEmailData)
	if err != nil {
		log.Error("メールのパースに失敗しました", zap.Error(err))
		response := createResponse("error", http.StatusInternalServerError, "Failed to parse email", messageID, err)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	logEmailData(emailData)

	if err := sendToExternalAPI(emailData, messageID); err != nil {
		log.Error("外部APIへの送信に失敗しました", zap.Error(err))
		response := createResponse("error", http.StatusInternalServerError, "Failed to send to external API", messageID, err)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	log.Info("メール処理が正常に完了しました", zap.String("messageId", messageID))
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

func sendToExternalAPI(emailData *models.EmailData, messageID string) error {
	log := logger.Logger

	payloadBytes, err := json.MarshalIndent(emailData, "", "  ")
	if err != nil {
		log.Error("ペイロードのJSONエンコードに失敗しました", zap.Error(err))
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	log.Info("外部APIにデータを送信します",
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
	if messageID != "" {
		req.Header.Set("X-Message-ID", messageID)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("HTTPリクエストの実行に失敗しました", zap.Error(err))
		return fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	log.Info("外部APIからレスポンスを受信しました",
		zap.String("messageId", messageID),
		zap.Int("statusCode", resp.StatusCode),
		zap.String("status", resp.Status),
	)

	if resp.StatusCode >= 400 { // 400以上をエラーとする
		logger.Logger.Error("外部APIがエラーを返しました",
			zap.String("messageId", messageID),
			zap.Int("statusCode", resp.StatusCode))
		return fmt.Errorf("external API returned error status: %d", resp.StatusCode)
	}

	logger.Logger.Info("外部APIにデータを送信しました",
		zap.String("messageId", messageID),
		zap.Int("statusCode", resp.StatusCode))
	return nil
}
