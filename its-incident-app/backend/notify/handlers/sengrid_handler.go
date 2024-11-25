// handlers/email_handler.go
package handlers

import (
	"net/http"
	"os"

	"notification/logger"

	"github.com/gin-gonic/gin"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

type SendLoginLinkRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Token     string `json:"token" binding:"required"`
	LoginURL  string `json:"login_url" binding:"required"`
	ExpiresIn string `json:"expires_in" binding:"required"`
}

func SendLoginLink(c *gin.Context) {
	logFields := []zap.Field{
		zap.String("handler", "SendLoginLink"),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	}

	// リクエストのバリデーション
	var req SendLoginLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Error("リクエストのバリデーションに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	logFields = append(logFields, zap.String("email", req.Email))

	// メール送信の準備
	from := mail.NewEmail(os.Getenv("EMAIL_FROM_NAME"), os.Getenv("EMAIL_FROM_ADDRESS"))
	to := mail.NewEmail("", req.Email)
	subject := "ログインリンク"

	plainTextContent := createPlainTextContent(req)
	htmlContent := createHTMLContent(req)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))

	// メール送信
	response, err := client.Send(message)
	if err != nil {
		logger.Logger.Error("メール送信に失敗しました",
			append(logFields,
				zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	// SendGridのレスポンス検証
	if response.StatusCode >= 300 {
		logger.Logger.Error("SendGridからエラーレスポンスを受信しました",
			append(logFields,
				zap.Int("status_code", response.StatusCode),
				zap.String("response_body", response.Body))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	logger.Logger.Info("ログインリンクメールを送信しました",
		append(logFields,
			zap.String("login_url", req.LoginURL),
			zap.String("expires_in", req.ExpiresIn))...)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login email sent successfully",
		"email":   req.Email,
	})
}

func createPlainTextContent(req SendLoginLinkRequest) string {
	return `以下のリンクからログインしてください：

` + req.LoginURL + `

このリンクは` + req.ExpiresIn + `で有効期限が切れます。`
}

func createHTMLContent(req SendLoginLinkRequest) string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>ログインリンク</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c3e50;">ログインリンク</h2>
        <p>以下のリンクからログインしてください：</p>
        <p style="margin: 20px 0;">
            <a href="` + req.LoginURL + `" 
               style="background-color: #3498db; 
                      color: white; 
                      padding: 10px 20px; 
                      text-decoration: none; 
                      border-radius: 5px; 
                      display: inline-block;">
                ログイン
            </a>
        </p>
        <p style="color: #7f8c8d; font-size: 0.9em;">
            このリンクは` + req.ExpiresIn + `で有効期限が切れます。
        </p>
        <p style="color: #7f8c8d; font-size: 0.8em;">
            このメールに心当たりがない場合は、無視していただいて構いません。
        </p>
    </div>
</body>
</html>`
}
