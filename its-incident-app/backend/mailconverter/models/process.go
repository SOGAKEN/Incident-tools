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
