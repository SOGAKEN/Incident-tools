package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"autopilot/logger"
	"autopilot/models"

	"go.uber.org/zap"
)

type AIService struct {
	endpoint    string
	token       string
	shortClient *http.Client
	longClient  *http.Client
}

const (
	defaultShortTimeout = 30 * time.Second
	defaultLongTimeout  = 90 * time.Second
)

func NewAIService(endpoint, token string) *AIService {
	service := &AIService{
		endpoint: endpoint,
		token:    token,
		shortClient: &http.Client{
			Timeout: defaultShortTimeout,
		},
		longClient: &http.Client{
			Timeout: defaultLongTimeout,
		},
	}

	logger.Logger.Info("AIサービスを初期化しました",
		zap.Bool("has_endpoint", endpoint != ""),
		zap.Bool("has_token", token != ""),
		zap.Duration("short_timeout", defaultShortTimeout),
		zap.Duration("long_timeout", defaultLongTimeout),
	)

	return service
}

func (s *AIService) ProcessEmail(ctx context.Context, emailData *models.EmailData) (*models.AIResponse, error) {
	if s.endpoint == "" {
		logger.Logger.Error("AIエンドポイントが設定されていません")
		return nil, fmt.Errorf("AI endpoint is not set")
	}

	if s.token == "" {
		logger.Logger.Error("AIトークンが設定されていません")
		return nil, fmt.Errorf("AI token is not set")
	}

	apiPayload := models.APIPayload{
		User: "system",
		Inputs: struct {
			Subject string `json:"subject"`
			From    string `json:"from"`
			Body    string `json:"body"`
		}{
			Subject: emailData.Subject,
			From:    emailData.From,
			Body:    emailData.Body,
		},
	}

	payloadBytes, err := json.Marshal(apiPayload)
	if err != nil {
		logger.Logger.Error("ペイロードのJSONエンコードに失敗しました",
			zap.Error(err),
			zap.String("subject", emailData.Subject),
		)
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	// リクエストペイロードはDEBUGレベル
	logger.Logger.Debug("AI APIリクエストペイロード",
		zap.String("payload", string(payloadBytes)),
	)

	req, err := http.NewRequestWithContext(ctx, "POST", s.endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Logger.Error("HTTPリクエストの作成に失敗しました",
			zap.Error(err),
			zap.String("endpoint", s.endpoint),
		)
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	// リクエスト送信情報はDEBUGレベル
	logger.Logger.Debug("AI APIにリクエストを送信します",
		zap.String("method", req.Method),
		zap.String("endpoint", req.URL.String()),
	)

	resp, err := s.longClient.Do(req)
	if err != nil {
		logger.Logger.Error("HTTPリクエストの実行に失敗しました",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Logger.Error("AI APIが異常なステータスを返しました",
			zap.Int("status_code", resp.StatusCode),
		)
		return nil, fmt.Errorf("AI API returned non-200 status: %d", resp.StatusCode)
	}

	var aiResponse models.AIResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResponse); err != nil {
		logger.Logger.Error("AIレスポンスのデコードに失敗しました",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to decode AI response: %v", err)
	}

	// バリデーション実行
	if err := s.ValidateResponse(&aiResponse); err != nil {
		logger.Logger.Error("AIレスポンスの検証に失敗しました",
			zap.Error(err),
			zap.Any("response", aiResponse),
		)
		return nil, fmt.Errorf("invalid AI response: %v", err)
	}

	// 処理完了のログは重要なのでINFOレベル
	logger.Logger.Info("AI処理が完了しました",
		zap.String("task_id", aiResponse.TaskID),
		zap.String("status", aiResponse.Data.Status),
	)

	return &aiResponse, nil
}

func (s *AIService) ValidateResponse(response *models.AIResponse) error {
	if response == nil {
		return fmt.Errorf("AI response is nil")
	}

	logFields := []zap.Field{
		zap.Bool("has_task_id", response.TaskID != ""),
		zap.Bool("has_status", response.Data.Status != ""),
		zap.Bool("has_error", response.Data.Error != nil),
	}

	// バリデーションエラーは重要なのでERRORレベル
	if response.TaskID == "" {
		logger.Logger.Error("AIレスポンスにtask_idが存在しません", logFields...)
		return fmt.Errorf("AI response missing task_id")
	}

	if response.Data.Status == "" {
		logger.Logger.Error("AIレスポンスにstatusが存在しません", logFields...)
		return fmt.Errorf("AI response missing status")
	}

	if response.Data.Error != nil {
		logger.Logger.Error("AIレスポンスにエラーが含まれています",
			append(logFields, zap.Any("error", response.Data.Error))...)
		return fmt.Errorf("AI processing error: %v", response.Data.Error)
	}

	// バリデーション完了はDEBUGレベル
	logger.Logger.Debug("AIレスポンスのバリデーションが完了しました", logFields...)
	return nil
}
