package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"autopilot/config"
	"autopilot/logger"
	"autopilot/models"

	"go.uber.org/zap"
)

type AIService struct {
	config      *config.AIConfig
	shortClient *http.Client
	longClient  *http.Client
	rand        *rand.Rand
}

func NewAIService(cfg *config.AIConfig) *AIService {
	source := rand.NewSource(time.Now().UnixNano())
	service := &AIService{
		config: cfg,
		shortClient: &http.Client{
			Timeout: cfg.ShortTimeout,
		},
		longClient: &http.Client{
			Timeout: cfg.LongTimeout,
		},
		rand: rand.New(source),
	}

	// 初期化時の設定情報は重要なのでINFOレベル
	logger.Logger.Info("AIサービスを初期化しました",
		zap.Duration("short_timeout", cfg.ShortTimeout),
		zap.Duration("long_timeout", cfg.LongTimeout),
		zap.Int("max_retries", cfg.MaxRetries))

	return service
}

func (s *AIService) ProcessEmail(ctx context.Context, emailData *models.EmailData) (*models.AIResponse, error) {
	// リクエストの開始をDEBUGレベルで記録
	logger.Logger.Debug("AI処理を開始します",
		zap.String("subject", emailData.Subject))

	var lastErr error
	var response *models.AIResponse

	for attempt := 1; attempt <= s.config.MaxRetries; attempt++ {
		response, lastErr = s.processEmailWithContext(ctx, emailData)
		if lastErr == nil {
			// 成功時はINFOレベル
			logger.Logger.Info("AI処理が完了しました",
				zap.String("task_id", response.TaskID),
				zap.Int("attempt", attempt))
			return response, nil
		}

		// 最後の試行ではリトライしない
		if attempt == s.config.MaxRetries {
			break
		}

		delay := s.calculateRetryDelay()
		if response == nil {
			response = models.NewErrorResponse("retry-"+fmt.Sprint(attempt), lastErr)
		}
		response.AddRetryInfo(attempt, delay, lastErr)

		// リトライはINFOレベル（想定内の動作）
		logger.Logger.Info("AI処理をリトライします",
			zap.Int("attempt", attempt),
			zap.Duration("delay", delay),
			zap.Error(lastErr))

		select {
		case <-ctx.Done():
			return response, fmt.Errorf("context cancelled during retry: %v", ctx.Err())
		case <-time.After(delay):
		}
	}

	// 全リトライ失敗はエラーレベル
	logger.Logger.Error("AI処理が失敗しました",
		zap.Error(lastErr),
		zap.Int("attempts", s.config.MaxRetries))

	if response == nil {
		response = models.NewErrorResponse("final-error", lastErr)
	}
	return response, fmt.Errorf("all retry attempts failed: %v", lastErr)
}

func (s *AIService) processEmailWithContext(ctx context.Context, emailData *models.EmailData) (*models.AIResponse, error) {
	if err := s.validateConfig(); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.Endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.Token)

	// HTTPリクエストの詳細はDEBUGレベル
	logger.Logger.Debug("AI APIリクエスト",
		zap.String("method", req.Method),
		zap.String("endpoint", s.config.Endpoint))

	resp, err := s.longClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI API returned non-200 status: %d", resp.StatusCode)
	}

	var aiResponse models.AIResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode AI response: %v", err)
	}

	if err := s.ValidateResponse(&aiResponse); err != nil {
		return nil, err
	}

	return &aiResponse, nil
}

func (s *AIService) ValidateResponse(response *models.AIResponse) error {
	if response == nil {
		return fmt.Errorf("AI response is nil")
	}

	if response.TaskID == "" {
		return fmt.Errorf("AI response missing task_id")
	}

	if response.Data.Status == "" {
		return fmt.Errorf("AI response missing status")
	}

	if response.Data.Error != nil {
		// エラー情報はERRORレベル
		logger.Logger.Error("AIレスポンスエラー",
			zap.Any("error", response.Data.Error),
			zap.String("task_id", response.TaskID))
		return fmt.Errorf("AI processing error: %v", response.Data.Error)
	}

	return nil
}

func (s *AIService) calculateRetryDelay() time.Duration {
	delta := int64(s.config.MaxRetryDelay - s.config.MinRetryDelay)
	if delta <= 0 {
		return s.config.MinRetryDelay
	}
	return s.config.MinRetryDelay + time.Duration(s.rand.Int63n(delta))
}

func (s *AIService) validateConfig() error {
	if s.config.Endpoint == "" {
		return fmt.Errorf("AI endpoint is not set")
	}
	if s.config.Token == "" {
		return fmt.Errorf("AI token is not set")
	}
	return nil
}
