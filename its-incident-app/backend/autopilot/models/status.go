package models

import "time"

// ProcessStatus は処理状態を表す型
type ProcessStatus string

const (
	// 処理状態の定義
	StatusPending  ProcessStatus = "pending"  // 処理待ち
	StatusRunning  ProcessStatus = "running"  // AI処理実行中
	StatusComplete ProcessStatus = "complete" // 処理完了
	StatusFailed   ProcessStatus = "failed"   // 処理失敗
)

// ProcessingStatus は処理の状態を表す構造体
type ProcessingStatus struct {
	MessageID   string        `json:"message_id"`
	Status      ProcessStatus `json:"status"`
	TaskID      string        `json:"task_id,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	Error       string        `json:"error,omitempty"`
}

// NewProcessingStatus は新しいProcessingStatusインスタンスを作成します
func NewProcessingStatus(messageID string) *ProcessingStatus {
	return &ProcessingStatus{
		MessageID: messageID,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
}

// SetRunning は状態を実行中に更新します
func (p *ProcessingStatus) SetRunning(taskID string) {
	p.Status = StatusRunning
	p.TaskID = taskID
}

// SetComplete は処理を完了状態に更新します
func (p *ProcessingStatus) SetComplete() {
	p.Status = StatusComplete
	now := time.Now()
	p.CompletedAt = &now
}

// SetFailed は処理を失敗状態に更新します
func (p *ProcessingStatus) SetFailed(err error) {
	p.Status = StatusFailed
	p.Error = err.Error()
	now := time.Now()
	p.CompletedAt = &now
}

// IsComplete は処理が完了しているかを確認します
func (p *ProcessingStatus) IsComplete() bool {
	return p.Status == StatusComplete
}

// IsFailed は処理が失敗しているかを確認します
func (p *ProcessingStatus) IsFailed() bool {
	return p.Status == StatusFailed
}

// IsFinished は処理が完了または失敗しているかを確認します
func (p *ProcessingStatus) IsFinished() bool {
	return p.IsComplete() || p.IsFailed()
}
