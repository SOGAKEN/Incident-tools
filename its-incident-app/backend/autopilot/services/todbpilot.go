package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"autopilot/logger"
	"autopilot/models"

	"go.uber.org/zap"
)

type DBPilotService struct {
	baseURL      string
	serviceToken string
	client       *http.Client
}

func NewDBPilotService(baseURL, serviceToken string) *DBPilotService {
	service := &DBPilotService{
		baseURL:      baseURL,
		serviceToken: serviceToken,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// 初期化時のログ
	logger.Logger.Info("DBPilotサービスを初期化しました",
		zap.Bool("has_base_url", baseURL != ""),
		zap.Bool("has_token", serviceToken != ""),
		zap.Duration("timeout", service.client.Timeout),
	)

	return service
}

func (s *DBPilotService) SaveEmail(emailData *models.EmailData, messageID string) error {
	startTime := time.Now()

	// 基本的なログフィールド
	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("operation", "SaveEmail"),
	}

	// EmailDataのデバッグログ
	if emailDataJSON, err := json.MarshalIndent(emailData, "", "  "); err == nil {
		logger.Logger.Debug("メールデータ",
			append(logFields, zap.String("email_data", string(emailDataJSON)))...)
	}

	// ペイロードの作成
	payload := models.EmailPayload{
		MessageID: messageID,
		EmailData: emailData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		logger.Logger.Error("ペイロードのJSONエンコードに失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to marshal email payload: %v", err)
	}

	// リクエストの作成
	req, err := s.createRequest("POST", "/emails", jsonData)
	if err != nil {
		logger.Logger.Error("リクエストの作成に失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to create request: %v", err)
	}

	// リクエスト情報のログ
	logger.Logger.Info("DBPilotへリクエストを送信します",
		append(logFields,
			zap.String("url", req.URL.String()),
			zap.String("method", req.Method),
			zap.Int("content_length", len(jsonData)))...)

	// リクエスト実行
	resp, err := s.client.Do(req)
	if err != nil {
		logger.Logger.Error("DBPilotへのリクエストに失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to save email to DBpilot: %v", err)
	}
	defer resp.Body.Close()

	// レスポンス処理
	respBody, _ := io.ReadAll(resp.Body)
	duration := time.Since(startTime)

	respFields := append(logFields,
		zap.Int("status_code", resp.StatusCode),
		zap.String("response_body", string(respBody)),
		zap.Duration("duration", duration))

	if resp.StatusCode != http.StatusOK {
		logger.Logger.Error("DBPilotがエラーを返しました", respFields...)
		return fmt.Errorf("failed to save email, status: %d, response: %s", resp.StatusCode, string(respBody))
	}

	logger.Logger.Info("メール保存が完了しました", respFields...)
	return nil
}

func (s *DBPilotService) SaveIncident(aiResponse *models.AIResponse, messageID string) error {
	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("operation", "SaveIncident"),
		zap.String("task_id", aiResponse.TaskID),
	}

	// APIRequestの形式に合わせてペイロードを構築
	payload := struct {
		TaskID        string `json:"task_id"`
		WorkflowRunID string `json:"workflow_run_id"`
		MessageID     string `json:"message_id"`
		Data          struct {
			ID         string `json:"id"`
			WorkflowID string `json:"workflow_id"`
			Status     string `json:"status"`
			Outputs    struct {
				Body         string               `json:"body"`
				User         string               `json:"user"`
				WorkflowLogs []models.WorkflowLog `json:"workflowLogs"`
				Host         string               `json:"host"`
				Priority     string               `json:"priority"`
				Subject      string               `json:"subject"`
				From         string               `json:"from"`
				Place        string               `json:"place"`
				Incident     string               `json:"incident"`
				Time         string               `json:"time"`
				IncidentID   int                  `json:"incidentID"`
				Judgment     string               `json:"judgment"`
				Sender       string               `json:"sender"`
				Final        string               `json:"final"`
			} `json:"outputs"`
			Error       interface{} `json:"error"`
			ElapsedTime float64     `json:"elapsed_time"`
			TotalTokens int         `json:"total_tokens"`
			TotalSteps  int         `json:"total_steps"`
			CreatedAt   int64       `json:"created_at"`
			FinishedAt  int64       `json:"finished_at"`
		} `json:"data"`
	}{
		TaskID:        aiResponse.TaskID,
		WorkflowRunID: aiResponse.WorkflowRunID,
		MessageID:     messageID,
		Data:          aiResponse.Data,
	}

	// デバッグログ
	if payloadJSON, err := json.MarshalIndent(payload, "", "  "); err == nil {
		logger.Logger.Debug("インシデントペイロード",
			append(logFields, zap.String("payload", string(payloadJSON)))...)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		logger.Logger.Error("インシデントペイロードのエンコードに失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to marshal incident payload: %v", err)
	}

	// 以下は既存のコード
	req, err := s.createRequest("POST", "/incidents", jsonData)
	if err != nil {
		logger.Logger.Error("インシデントリクエストの作成に失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to create request: %v", err)
	}

	logger.Logger.Info("インシデントデータを送信します",
		append(logFields,
			zap.String("url", req.URL.String()),
			zap.String("method", req.Method),
			zap.Int("content_length", len(jsonData)))...)

	resp, err := s.client.Do(req)
	if err != nil {
		logger.Logger.Error("インシデントの送信に失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to send incident to DBpilot: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取り
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.Logger.Error("インシデント保存でエラーが発生しました",
			append(logFields,
				zap.Int("status_code", resp.StatusCode),
				zap.String("response_body", string(respBody)))...)
		return fmt.Errorf("failed to save incident, status: %d, response: %s", resp.StatusCode, string(respBody))
	}

	logger.Logger.Info("インシデントの保存が完了しました",
		append(logFields,
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", string(respBody)))...)
	return nil
}

func (s *DBPilotService) createRequest(method, path string, payload []byte) (*http.Request, error) {
	if s.baseURL == "" {
		logger.Logger.Error("DBPilot URLが設定されていません")
		return nil, fmt.Errorf("DBPilot URL is not set")
	}

	if s.serviceToken == "" {
		logger.Logger.Error("サービストークンが設定されていません")
		return nil, fmt.Errorf("service token is not set")
	}

	req, err := http.NewRequest(method, s.baseURL+path, bytes.NewBuffer(payload))
	if err != nil {
		logger.Logger.Error("HTTPリクエストの作成に失敗しました",
			zap.Error(err),
			zap.String("method", method),
			zap.String("path", path))
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.serviceToken)

	return req, nil
}

// service/todbpilot.go に追加

// GetProcessingStatus は処理状態を取得します
func (s *DBPilotService) GetProcessingStatus(messageID string) (*models.ProcessingStatus, error) {
	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("operation", "GetProcessingStatus"),
	}

	// リクエストの作成
	req, err := s.createRequest("GET", fmt.Sprintf("/status/%s", messageID), nil)
	if err != nil {
		logger.Logger.Error("リクエストの作成に失敗しました",
			append(logFields, zap.Error(err))...)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// リクエスト情報のログ
	logger.Logger.Info("処理状態を確認します",
		append(logFields,
			zap.String("url", req.URL.String()),
			zap.String("method", req.Method))...)

	// リクエスト実行
	resp, err := s.client.Do(req)
	if err != nil {
		logger.Logger.Error("処理状態の取得に失敗しました",
			append(logFields, zap.Error(err))...)
		return nil, fmt.Errorf("failed to get processing status: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスの処理
	if resp.StatusCode == http.StatusNotFound {
		logger.Logger.Info("指定されたメッセージIDの処理状態が見つかりません", logFields...)
		return nil, fmt.Errorf("processing status not found for message_id: %s", messageID)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		logger.Logger.Error("処理状態の取得でエラーが発生しました",
			append(logFields,
				zap.Int("status_code", resp.StatusCode),
				zap.String("response_body", string(respBody)))...)
		return nil, fmt.Errorf("failed to get processing status, status: %d, response: %s",
			resp.StatusCode, string(respBody))
	}

	// レスポンスのデコード
	var status models.ProcessingStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		logger.Logger.Error("レスポンスのデコードに失敗しました",
			append(logFields, zap.Error(err))...)
		return nil, fmt.Errorf("failed to decode processing status: %v", err)
	}

	logger.Logger.Info("処理状態を取得しました",
		append(logFields,
			zap.String("status", string(status.Status)),
			zap.String("task_id", status.TaskID))...)

	return &status, nil
}

// UpdateProcessingStatus は処理状態を更新します
func (s *DBPilotService) UpdateProcessingStatus(status *models.ProcessingStatus) error {
	logFields := []zap.Field{
		zap.String("message_id", status.MessageID),
		zap.String("operation", "UpdateProcessingStatus"),
		zap.String("status", string(status.Status)),
	}

	// ペイロードの作成
	jsonData, err := json.Marshal(status)
	if err != nil {
		logger.Logger.Error("ステータスのJSONエンコードに失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to marshal status: %v", err)
	}

	// リクエストの作成
	req, err := s.createRequest("PUT", fmt.Sprintf("/status/%s", status.MessageID), jsonData)
	if err != nil {
		logger.Logger.Error("リクエストの作成に失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to create request: %v", err)
	}

	// リクエスト実行
	logger.Logger.Info("処理状態を更新します", logFields...)

	resp, err := s.client.Do(req)
	if err != nil {
		logger.Logger.Error("処理状態の更新に失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to update processing status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		logger.Logger.Error("処理状態の更新でエラーが発生しました",
			append(logFields,
				zap.Int("status_code", resp.StatusCode),
				zap.String("response_body", string(respBody)))...)
		return fmt.Errorf("failed to update status, status: %d, response: %s",
			resp.StatusCode, string(respBody))
	}

	logger.Logger.Info("処理状態を更新しました", logFields...)
	return nil
}
