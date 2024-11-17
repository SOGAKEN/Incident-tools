package models

// EmailData はメールのデータ構造を定義します
type EmailData struct {
	From                    string `json:"from"`
	To                      string `json:"to"`
	Subject                 string `json:"subject"`
	Date                    string `json:"date"`
	OriginalMessageID       string `json:"original_message_id"`
	MIMEVersion             string `json:"mime_version"`
	ContentType             string `json:"content_type"`
	ContentTransferEncoding string `json:"content_transfer_encoding"`
	CC                      string `json:"cc"`
	Body                    string `json:"body"`
	FileName                string `json:"file_name,omitempty"`
}

// APIResponse はAPIレスポンスの構造を定義します
type APIResponse struct {
	Status    string     `json:"status"`            // "success" or "error"
	Code      int        `json:"code"`              // HTTPステータスコード
	Message   string     `json:"message,omitempty"` // 処理結果の説明
	TraceID   string     `json:"trace_id"`          // X-Message-IDの値
	Timestamp string     `json:"timestamp"`         // 処理時のタイムスタンプ
	Error     *ErrorInfo `json:"error,omitempty"`   // エラー情報（エラー時のみ）
}

// ErrorInfo はエラー詳細情報の構造を定義します
type ErrorInfo struct {
	Type    string `json:"type"`             // エラーの種類（parse_error, api_error, etc.）
	Message string `json:"message"`          // エラーメッセージ
	Detail  string `json:"detail,omitempty"` // 詳細なエラー情報
}
