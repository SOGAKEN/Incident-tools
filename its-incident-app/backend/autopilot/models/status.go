package models

import "time"

// ProcessStatus は処理状態を表す型です
type ProcessStatus string

const (
	// 処理状態の定義
	StatusPending  ProcessStatus = "PENDING"  // 処理待ち
	StatusRunning  ProcessStatus = "RUNNING"  // 処理実行中
	StatusComplete ProcessStatus = "COMPLETE" // 処理完了
	StatusFailed   ProcessStatus = "FAILED"   // 処理失敗
)

// ServiceType はマイクロサービスの種類を表す型です
type ServiceType string

const (
	ServiceMailConverter ServiceType = "mail-converter"
	ServiceAutoPilot     ServiceType = "auto-pilot"
	ServiceDBPilot       ServiceType = "db-pilot"
)

// ProcessingStatus は基本的な処理状態を表す構造体です
// DBPilotとの互換性のために維持されています
type ProcessingStatus struct {
	MessageID   string        `json:"message_id" datastore:"message_id"`
	Status      ProcessStatus `json:"status" datastore:"status"`
	TaskID      string        `json:"task_id,omitempty" datastore:"task_id,omitempty"`
	CreatedAt   time.Time     `json:"created_at" datastore:"created_at"`
	CompletedAt *time.Time    `json:"completed_at,omitempty" datastore:"completed_at,omitempty"`
	Error       string        `json:"error,omitempty" datastore:"error,omitempty"`
}

// EmailProcessing はメール処理の全体状態を管理する拡張構造体です
type EmailProcessing struct {
	MessageID    string        `datastore:"-" json:"message_id"`
	Status       ProcessStatus `datastore:"status" json:"status"`
	CreatedAt    time.Time     `datastore:"created_at" json:"created_at"`
	UpdatedAt    time.Time     `datastore:"updated_at" json:"updated_at"`
	CompletedAt  *time.Time    `datastore:"completed_at,omitempty" json:"completed_at,omitempty"`
	ErrorMessage string        `datastore:"error_message,omitempty" json:"error_message,omitempty"`
}

// ServiceState は各マイクロサービスの詳細な状態を管理する構造体です
type ServiceState struct {
	MessageID    string        `datastore:"-" json:"message_id"`
	ServiceType  ServiceType   `datastore:"service_type" json:"service_type"`
	Status       ProcessStatus `datastore:"status" json:"status"`
	TaskID       string        `datastore:"task_id,omitempty" json:"task_id,omitempty"`
	ErrorCode    string        `datastore:"error_code,omitempty" json:"error_code,omitempty"`
	ErrorMessage string        `datastore:"error_message,omitempty" json:"error_message,omitempty"`
	//	AIResponse   *AIResponse   `datastore:"ai_response,omitempty" json:"ai_response,omitempty"`
	CreatedAt   time.Time  `datastore:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `datastore:"updated_at" json:"updated_at"`
	CompletedAt *time.Time `datastore:"completed_at,omitempty" json:"completed_at,omitempty"`
}

// ProcessingStatusのメソッド群
func NewProcessingStatus(messageID string) *ProcessingStatus {
	return &ProcessingStatus{
		MessageID: messageID,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
}

func (p *ProcessingStatus) SetRunning(taskID string) {
	p.Status = StatusRunning
	p.TaskID = taskID
}

func (p *ProcessingStatus) SetComplete() {
	p.Status = StatusComplete
	now := time.Now()
	p.CompletedAt = &now
}

func (p *ProcessingStatus) SetFailed(err error) {
	p.Status = StatusFailed
	p.Error = err.Error()
	now := time.Now()
	p.CompletedAt = &now
}

func (p *ProcessingStatus) IsComplete() bool {
	return p.Status == StatusComplete
}

func (p *ProcessingStatus) IsFailed() bool {
	return p.Status == StatusFailed
}

func (p *ProcessingStatus) IsFinished() bool {
	return p.IsComplete() || p.IsFailed()
}

// EmailProcessingのメソッド群
func (e *EmailProcessing) SetRunning() {
	e.Status = StatusRunning
	e.UpdatedAt = time.Now()
}

func (e *EmailProcessing) SetComplete() {
	e.Status = StatusComplete
	now := time.Now()
	e.CompletedAt = &now
	e.UpdatedAt = now
}

func (e *EmailProcessing) SetError(errorMessage string) {
	e.Status = StatusFailed
	e.ErrorMessage = errorMessage
	now := time.Now()
	e.CompletedAt = &now
	e.UpdatedAt = now
}

// ServiceStateのメソッド群
func (s *ServiceState) SetRunning(taskID string) {
	s.Status = StatusRunning
	s.TaskID = taskID
	s.UpdatedAt = time.Now()
}

func (s *ServiceState) SetComplete(aiResponse *AIResponse) {
	s.Status = StatusComplete
	//s.AIResponse = aiResponse
	now := time.Now()
	s.CompletedAt = &now
	s.UpdatedAt = now
}

func (s *ServiceState) SetError(errorCode, errorMessage string) {
	s.Status = StatusFailed
	s.ErrorCode = errorCode
	s.ErrorMessage = errorMessage
	now := time.Now()
	s.CompletedAt = &now
	s.UpdatedAt = now
}
