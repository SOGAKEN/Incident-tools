package models

import "time"

type ProcessingStatus string

const (
	StatusPending  ProcessingStatus = "PENDING"
	StatusRunning  ProcessingStatus = "RUNNING"
	StatusComplete ProcessingStatus = "COMPLETE"
	StatusError    ProcessingStatus = "ERROR"
)

// EmailProcessing メール処理の全体状態
type EmailProcessing struct {
	MessageID string           `datastore:"-"` // キーとして使用
	Status    ProcessingStatus `datastore:"status"`
	CreatedAt time.Time        `datastore:"created_at"`
	UpdatedAt time.Time        `datastore:"updated_at"`
}

// ServiceState 各サービスの状態
type ServiceState struct {
	MessageID    string           `datastore:"-"` // キーとして使用
	ServiceType  string           `datastore:"service_type"`
	Status       ProcessingStatus `datastore:"status"`
	ErrorCode    string           `datastore:"error_code,omitempty"`
	ErrorMessage string           `datastore:"error_message,omitempty"`
	EmailData    *EmailData       `datastore:"email_data,omitempty"`
	CreatedAt    time.Time        `datastore:"created_at"`
	UpdatedAt    time.Time        `datastore:"updated_at"`
}

const MaxBodySize = 1000 // 本文の最大バイト数

// ServiceStateForStore はDatastore保存用に本文を制限したServiceStateのコピーを作成します
func (s *ServiceState) ServiceStateForStore() *ServiceState {
	copy := *s // 浅いコピー

	if s.EmailData != nil {
		// EmailDataのディープコピー
		emailDataCopy := *s.EmailData
		copy.EmailData = &emailDataCopy

		// 保存用に本文を制限
		if len(copy.EmailData.Body) > MaxBodySize {
			for i := range copy.EmailData.Body {
				if i >= MaxBodySize {
					copy.EmailData.Body = copy.EmailData.Body[:i] + "... (truncated)"
					break
				}
			}
		}
	}

	return &copy
}