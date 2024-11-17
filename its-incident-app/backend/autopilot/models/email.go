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

// EmailPayload はDBpilotのemailsエンドポイントへ送信するペイロードです
type EmailPayload struct {
	MessageID string     `json:"message_id"`
	EmailData *EmailData `json:"email_data"`
}

// APIPayload は外部APIへのリクエストペイロードの構造を定義します
type APIPayload struct {
	Inputs struct {
		Subject string `json:"subject"`
		From    string `json:"from"`
		Body    string `json:"body"`
	} `json:"inputs"`
	User string `json:"user"`
}
