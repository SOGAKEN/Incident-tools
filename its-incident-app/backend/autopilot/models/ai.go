package models

import "encoding/json"

// WorkflowLog はワークフローのログ情報を定義します
type WorkflowLog struct {
	Answers map[string]string // answer1, answer2, ... を格納
}

// AIResponse は外部APIからのレスポンスを定義します
type AIResponse struct {
	TaskID        string `json:"task_id"`
	WorkflowRunID string `json:"workflow_run_id"`
	Data          struct {
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
	} `json:"data"`
}

// AIResponsePayload はDBpilotのincidentsエンドポイントへ送信するペイロードです
type AIResponsePayload struct {
	MessageID  string      `json:"message_id"`
	AIResponse *AIResponse `json:"ai_response"`
}
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
