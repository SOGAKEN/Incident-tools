package models

import (
	"encoding/json"
	"time"
)

// WorkflowLog はワークフローのログ情報を定義します
type WorkflowLog map[string]string

// AIResponseData は処理結果のデータ構造を定義します
type AIResponseData struct {
	ID         string `json:"id"`
	WorkflowID string `json:"workflow_id"`
	Status     string `json:"status"`
	Outputs    struct {
		Body         string        `json:"body"`
		User         string        `json:"user"`
		WorkflowLogs []WorkflowLog `json:"workflowLogs"`
		Host         string        `json:"host"`
		Priority     string        `json:"priority"`
		Subject      string        `json:"subject"`
		From         string        `json:"from"`
		Place        string        `json:"place"`
		Incident     string        `json:"incident"`
		Time         string        `json:"time"`
		IncidentID   int           `json:"incidentID"`
		Judgment     string        `json:"judgment"`
		Sender       string        `json:"sender"`
		Final        string        `json:"final"`
	} `json:"outputs"`
	Error       interface{} `json:"error"`
	ElapsedTime float64     `json:"elapsed_time"`
	TotalTokens int         `json:"total_tokens"`
	TotalSteps  int         `json:"total_steps"`
	CreatedAt   int64       `json:"created_at"`
	FinishedAt  int64       `json:"finished_at"`
}

// AIResponse は外部APIからのレスポンスを定義します
type AIResponse struct {
	TaskID        string         `json:"task_id"`
	WorkflowRunID string         `json:"workflow_run_id"`
	Data          AIResponseData `json:"data"`
}

// AIResponsePayload はDBpilotのincidentsエンドポイントへ送信するペイロードです
type AIResponsePayload struct {
	MessageID  string      `json:"message_id"`
	AIResponse *AIResponse `json:"ai_response"`
}

// OutputsData は生のワークフローログを持つ出力データを定義します
type OutputsData struct {
	Body         string          `json:"body"`
	User         string          `json:"user"`
	WorkflowLogs json.RawMessage `json:"workflowLogs"`
	Host         string          `json:"host"`
	Priority     string          `json:"priority"`
	Subject      string          `json:"subject"`
	From         string          `json:"from"`
	Place        string          `json:"place"`
	Incident     string          `json:"incident"`
	Time         string          `json:"time"`
	IncidentID   int             `json:"incidentID"`
	Judgment     string          `json:"judgment"`
	Sender       string          `json:"sender"`
	Final        string          `json:"final"`
}

// NewErrorResponse はエラー情報を含むAIResponseを生成するヘルパー関数です
func NewErrorResponse(messageID string, err error) *AIResponse {
	now := time.Now()
	unixNow := now.Unix()

	response := &AIResponse{
		TaskID:        "error-" + messageID,
		WorkflowRunID: "error-workflow-" + messageID,
		Data: AIResponseData{
			ID:         "error-" + messageID,
			WorkflowID: "error-workflow-" + messageID,
			Status:     "error",
			Error:      err.Error(),
			Outputs: struct {
				Body         string        `json:"body"`
				User         string        `json:"user"`
				WorkflowLogs []WorkflowLog `json:"workflowLogs"`
				Host         string        `json:"host"`
				Priority     string        `json:"priority"`
				Subject      string        `json:"subject"`
				From         string        `json:"from"`
				Place        string        `json:"place"`
				Incident     string        `json:"incident"`
				Time         string        `json:"time"`
				IncidentID   int           `json:"incidentID"`
				Judgment     string        `json:"judgment"`
				Sender       string        `json:"sender"`
				Final        string        `json:"final"`
			}{
				Body:     err.Error(),
				Priority: "high",
				Time:     now.Format(time.RFC3339),
				Final:    "error",
				WorkflowLogs: []WorkflowLog{
					{
						"step":    "1",
						"action":  "error",
						"message": err.Error(),
						"time":    now.Format(time.RFC3339),
					},
				},
			},
			ElapsedTime: 0,
			TotalTokens: 0,
			TotalSteps:  1,
			CreatedAt:   unixNow,
			FinishedAt:  unixNow,
		},
	}

	return response
}

// IsError はレスポンスがエラー状態かどうかを判定します
func (r *AIResponse) IsError() bool {
	return r.Data.Status == "error" || r.Data.Error != nil
}

// GetError はエラーメッセージを取得します
func (r *AIResponse) GetError() string {
	if r.Data.Error == nil {
		return ""
	}

	switch e := r.Data.Error.(type) {
	case string:
		return e
	case error:
		return e.Error()
	default:
		if jsonBytes, err := json.Marshal(e); err == nil {
			return string(jsonBytes)
		}
		return "unknown error"
	}
}
