package models

import (
	"encoding/json"
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

type User struct {
	BaseModel
	Email    string `gorm:"unique"`
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
	Datetime  time.Time `gorm:"not null"`
	Status    string    `gorm:"size:50;not null"`
	Assignee  string    `gorm:"size:100;not null"`
	Vender    int
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
	Datetime   time.Time `gorm:"not null"`
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

	Body         string `gorm:"type:text"`
	User         string `gorm:"size:100"`
	WorkflowLogs string `gorm:"type:jsonb"`
	Host         string `gorm:"size:100"`
	Priority     string `gorm:"size:50"`
	Subject      string `gorm:"size:200"`
	From         string `gorm:"size:100"`
	Place        string `gorm:"size:200"`
	IncidentText string `gorm:"type:text"`
	Time         string `gorm:"size:50"`
	Judgment     string `gorm:"size:100"`
	Sender       string `gorm:"size:100"`
	Final        string `gorm:"type:text"`

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
	RawJSON       string `gorm:"type:jsonb"`
}

type EmailData struct {
	BaseModel
	MessageID               string `json:"message_id" gorm:"type:varchar(255);not null;uniqueIndex"` // PayloadのメッセージID
	EmailFrom               string `json:"from" gorm:"type:varchar(255);not null"`                   // 差出人
	To                      string `json:"to" gorm:"type:varchar(255);not null"`                     // 宛先
	Subject                 string `json:"subject" gorm:"type:varchar(255)"`                         // 件名
	Date                    string `json:"date" gorm:"type:varchar(255)"`                            // メールの日付
	OriginalMessageID       string `json:"original_message_id" gorm:"type:varchar(255)"`             // メッセージID
	MIMEVersion             string `json:"mime_version" gorm:"type:varchar(50)"`                     // MIMEバージョン
	ContentType             string `json:"content_type" gorm:"type:varchar(255)"`                    // コンテンツタイプ
	ContentTransferEncoding string `json:"content_transfer_encoding" gorm:"type:varchar(50)"`        // コンテンツ転送エンコーディング
	CC                      string `json:"cc" gorm:"type:varchar(255)"`                              // CC
	Body                    string `json:"body" gorm:"type:text"`                                    // メール本文
	FileName                string `json:"file_name,omitempty" gorm:"type:varchar(255)"`             // ファイル名（添付ファイル）
}

type EmailPayload struct {
	MessageID string     `json:"message_id"`
	EmailData *EmailData `json:"email_data"`
}

// models/models.go

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

// DB操作のためのメソッド群

// DeleteSessionByEmail はメールアドレスに基づいてセッションを削除
func DeleteSessionByEmail(db *gorm.DB, email string) error {
	result := db.Where("email = ?", email).Delete(&LoginSession{})
	return result.Error
}

// CreateSession は新しいセッションを作成
func CreateSession(db *gorm.DB, session *LoginSession) error {
	return db.Create(session).Error
}

// GetSessionByEmail はメールアドレスに基づいてセッションを取得
func GetSessionByEmail(db *gorm.DB, email string) (*LoginSession, error) {
	var session LoginSession
	err := db.Where("email = ?", email).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetUserByEmail はメールアドレスに基づいてユーザーを取得
func GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser は新しいユーザーを作成
func CreateUser(db *gorm.DB, user *User) error {
	return db.Create(user).Error
}

// UpdateUser は既存のユーザー情報を更新
func UpdateUser(db *gorm.DB, user *User) error {
	return db.Save(user).Error
}