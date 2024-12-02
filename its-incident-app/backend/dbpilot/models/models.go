package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// BaseModel は共通のフィールドを持つ基本モデル
type BaseModel struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone"`
}

// BeforeCreate は作成時に東京時間を設定
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(jst)
	b.CreatedAt = now
	b.UpdatedAt = now
	return nil
}

// BeforeUpdate は更新時に東京時間を設定
func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	b.UpdatedAt = time.Now().In(jst)
	return nil
}

// IncidentStatus はインシデントの状態を管理するモデル
type IncidentStatus struct {
	BaseModel
	Code         int    `gorm:"uniqueIndex;not null;comment:システムで使用する状態コード"`
	Name         string `gorm:"size:50;not null;comment:画面表示用の状態名"`
	Description  string `gorm:"type:text;comment:状態の詳細説明"`
	IsActive     bool   `gorm:"default:true;comment:状態の有効/無効を管理"`
	DisplayOrder int    `gorm:"comment:画面表示時の並び順"`

	// リレーション
	Incidents []Incident `gorm:"foreignKey:StatusID"`
}

// IncidentStatusTransition はステータス遷移を管理するモデル
type IncidentStatusTransition struct {
	BaseModel
	FromStatusID uint           `gorm:"not null"`
	ToStatusID   uint           `gorm:"not null"`
	FromStatus   IncidentStatus `gorm:"foreignKey:FromStatusID"`
	ToStatus     IncidentStatus `gorm:"foreignKey:ToStatusID"`
	Allowed      bool           `gorm:"default:true"`

	UniqueTransition string `gorm:"uniqueIndex:idx_status_transition"`
}

// BeforeCreate は遷移の一意性を確保するためのフック
func (ist *IncidentStatusTransition) BeforeCreate(tx *gorm.DB) error {
	ist.UniqueTransition = fmt.Sprintf("%d-%d", ist.FromStatusID, ist.ToStatusID)
	return nil
}

type User struct {
	BaseModel
	Email    string `gorm:"unique;type:varchar(255);not null"`
	Password string
	Profile  Profile `gorm:"foreignKey:UserID"`
}

type Profile struct {
	BaseModel
	UserID   uint `gorm:"unique"`
	Name     string
	ImageURL string
}

type LoginSession struct {
	BaseModel
	UserID    uint
	Email     string
	SessionID string `gorm:"unique"`
	ExpiresAt time.Time
}

type Incident struct {
	BaseModel
	Datetime  time.Time      `gorm:"not null"`
	StatusID  uint           `gorm:"not null;comment:状態ID"`
	Status    IncidentStatus `gorm:"foreignKey:StatusID"`
	Assignee  string         `gorm:"size:100;not null"`
	Vender    int
	MessageID string             `gorm:"size:100"`
	Responses []Response         `gorm:"foreignKey:IncidentID"`
	Relations []IncidentRelation `gorm:"foreignKey:IncidentID"`
	APIData   APIResponseData    `gorm:"foreignKey:IncidentID"`
}

type IncidentRelation struct {
	BaseModel
	IncidentID        uint     `gorm:"not null"`
	RelatedIncident   Incident `gorm:"foreignKey:RelatedIncidentID"`
	RelatedIncidentID uint     `gorm:"not null"`
}

type Response struct {
	BaseModel
	IncidentID uint      `gorm:"not null"`
	Datetime   time.Time `gorm:"type:timestamp with time zone;not null"`
	Responder  string    `gorm:"size:100;not null"`
	Content    string    `gorm:"type:text;not null"`
}

type APIResponseData struct {
	BaseModel
	IncidentID    uint   `gorm:"uniqueIndex"`
	TaskID        string `gorm:"size:100"`
	WorkflowRunID string `gorm:"size:100"`
	WorkflowID    string `gorm:"size:100"`
	Status        string `gorm:"size:50"`

	Body           string `gorm:"type:text"`
	User           string `gorm:"size:100"`
	WorkflowLogs   string `gorm:"type:jsonb"`
	Host           string `gorm:"size:100"`
	Priority       string `gorm:"size:50"`
	Subject        string `gorm:"size:200"`
	From           string `gorm:"size:100"`
	Place          string `gorm:"size:200"`
	IncidentText   string `gorm:"type:text"`
	Time           string `gorm:"size:50"`
	Judgment       string `gorm:"size:100"`
	Sender         string `gorm:"size:100"`
	Final          string `gorm:"type:text"`
	IncidentNumber int

	ElapsedTime float64
	TotalTokens int
	TotalSteps  int
	CreatedAt   int64
	FinishedAt  int64
	Error       string `gorm:"type:text"`
	RawResponse string `gorm:"type:jsonb"`
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

type APIRequest struct {
	TaskID        string `json:"task_id"`
	WorkflowRunID string `json:"workflow_run_id"`
	MessageID     string `json:"message_id"`
	Data          struct {
		ID          string      `json:"id"`
		WorkflowID  string      `json:"workflow_id"`
		Status      string      `json:"status"`
		Outputs     OutputsData `json:"outputs"`
		Error       interface{} `json:"error"`
		ElapsedTime float64     `json:"elapsed_time"`
		TotalTokens int         `json:"total_tokens"`
		TotalSteps  int         `json:"total_steps"`
		CreatedAt   int64       `json:"created_at"`
		FinishedAt  int64       `json:"finished_at"`
	} `json:"data"`
}

type ErrorLog struct {
	BaseModel
	TaskID        string `gorm:"size:100"`
	WorkflowRunID string `gorm:"size:100"`
	WorkflowID    string `gorm:"size:100"`
	Status        string `gorm:"size:50"`
	MessageID     string `gorm:"size:100"`
	RawJSON       string `gorm:"type:jsonb"`
}

type EmailData struct {
	BaseModel
	MessageID               string `json:"message_id" gorm:"type:varchar(255);not null;uniqueIndex"`
	EmailFrom               string `json:"from" gorm:"type:varchar(255);not null"`
	To                      string `json:"to" gorm:"type:varchar(255);not null"`
	Subject                 string `json:"subject" gorm:"type:varchar(255)"`
	Date                    string `json:"date" gorm:"type:varchar(255)"`
	OriginalMessageID       string `json:"original_message_id" gorm:"type:varchar(255)"`
	MIMEVersion             string `json:"mime_version" gorm:"type:varchar(50)"`
	ContentType             string `json:"content_type" gorm:"type:varchar(255)"`
	ContentTransferEncoding string `json:"content_transfer_encoding" gorm:"type:varchar(50)"`
	CC                      string `json:"cc" gorm:"type:varchar(255)"`
	Body                    string `json:"body" gorm:"type:text"`
	FileName                string `json:"file_name,omitempty" gorm:"type:varchar(255)"`
}

type EmailPayload struct {
	MessageID string     `json:"message_id"`
	EmailData *EmailData `json:"email_data"`
}

type APIResponseDataQuery struct {
	IncidentID    *uint   `json:"incident_id,omitempty"`
	TaskID        *string `json:"task_id,omitempty"`
	WorkflowRunID *string `json:"workflow_run_id,omitempty"`
	WorkflowID    *string `json:"workflow_id,omitempty"`
	Status        *string `json:"status,omitempty"`

	Body           *string `json:"body,omitempty"`
	User           *string `json:"user,omitempty"`
	Host           *string `json:"host,omitempty"`
	Priority       *string `json:"priority,omitempty"`
	Subject        *string `json:"subject,omitempty"`
	From           *string `json:"from,omitempty"`
	Place          *string `json:"place,omitempty"`
	IncidentText   *string `json:"incident_text,omitempty"`
	Time           *string `json:"time,omitempty"`
	Judgment       *string `json:"judgment,omitempty"`
	Sender         *string `json:"sender,omitempty"`
	Final          *string `json:"final,omitempty"`
	IncidentNumber *int    `json:"incident_number"`

	ElapsedTimeMin *float64 `json:"elapsed_time_min,omitempty"`
	ElapsedTimeMax *float64 `json:"elapsed_time_max,omitempty"`
	TotalTokensMin *int     `json:"total_tokens_min,omitempty"`
	TotalTokensMax *int     `json:"total_tokens_max,omitempty"`
	TotalStepsMin  *int     `json:"total_steps_min,omitempty"`
	TotalStepsMax  *int     `json:"total_steps_max,omitempty"`

	CreatedAtStart  *int64 `json:"created_at_start,omitempty"`
	CreatedAtEnd    *int64 `json:"created_at_end,omitempty"`
	FinishedAtStart *int64 `json:"finished_at_start,omitempty"`
	FinishedAtEnd   *int64 `json:"finished_at_end,omitempty"`

	Limit         *int    `json:"limit,omitempty"`
	Offset        *int    `json:"offset,omitempty"`
	SortBy        *string `json:"sort_by,omitempty"`
	SortDirection *string `json:"sort_direction,omitempty"`
}

type ProcessStatus string

const (
	StatusPending  ProcessStatus = "pending"
	StatusRunning  ProcessStatus = "running"
	StatusComplete ProcessStatus = "complete"
	StatusFailed   ProcessStatus = "failed"
)

type ProcessingStatus struct {
	gorm.Model
	MessageID   string        `gorm:"uniqueIndex" json:"message_id"`
	Status      ProcessStatus `gorm:"type:varchar(20)" json:"status"`
	TaskID      string        `json:"task_id,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	Error       string        `json:"error,omitempty"`
}

type LoginToken struct {
	gorm.Model
	Email     string    `gorm:"type:varchar(255);index;not null"`
	Token     string    `gorm:"uniqueIndex;type:varchar(255);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsExpired bool      `gorm:"default:false"`
}

type TokenAccess struct {
	BaseModel
	TokenID    uint      `gorm:"index;not null"`
	Token      string    `gorm:"type:varchar(255);not null"`
	Email      string    `gorm:"type:varchar(255);not null"`
	IP         string    `gorm:"type:varchar(255);not null"`
	UserAgent  string    `gorm:"type:text"`
	AccessedAt time.Time `gorm:"type:timestamp with time zone;not null"`
}

type LoginTokenRequest struct {
	Email     string    `json:"email" binding:"required,email"`
	Token     string    `json:"token" binding:"required"`
	ExpiresAt time.Time `json:"expires_at" binding:"required"`
}

type TokenVerificationResponse struct {
	Email    string `json:"email"`
	UserID   uint   `json:"user_id"`
	Name     string `json:"name,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}
