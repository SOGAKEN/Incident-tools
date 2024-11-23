// handlers/account_handler.go
package handlers

import (
	"auth/logger"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type CreateAccountRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type DBPilotAccountRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func CreateAccount(c *gin.Context) {
	logFields := []zap.Field{
		zap.String("handler", "CreateAccount"),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	}

	// リクエストのバリデーション
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Error("リクエストのバリデーションに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	logFields = append(logFields,
		zap.String("email", req.Email),
		zap.String("name", req.Name))

	logger.Logger.Info("アカウント作成を開始します", logFields...)

	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Logger.Error("パスワードのハッシュ化に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// DBPilotへのリクエスト準備
	dbPilotReq := DBPilotAccountRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	jsonData, err := json.Marshal(dbPilotReq)
	if err != nil {
		logger.Logger.Error("リクエストのJSONエンコードに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare request"})
		return
	}

	bearerToken := os.Getenv("SERVICE_TOKEN")
	if bearerToken == "" {
		logger.Logger.Error("Bearer tokenが設定されていません",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}
	// DB Pilotへのリクエスト作成
	dbPilotURL := os.Getenv("DB_PILOT_SERVICE_URL") + "/accounts"
	request, err := http.NewRequest("POST", dbPilotURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Logger.Error("DBPilotへのリクエスト作成に失敗しました",
			append(logFields,
				zap.String("url", dbPilotURL),
				zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// ヘッダーの設定
	token := "Bearer " + bearerToken
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", token)

	// リクエストの送信
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Logger.Error("DBPilotへのリクエスト送信に失敗しました",
			append(logFields,
				zap.String("url", dbPilotURL),
				zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request"})
		return
	}
	defer resp.Body.Close()

	// レスポンスの確認
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logger.Logger.Error("DBPilotからエラーレスポンスを受信しました",
			append(logFields,
				zap.Int("status_code", resp.StatusCode),
				zap.String("response_body", string(body)))...)

		// メールアドレスの重複エラーの場合は専用のエラーメッセージを返す
		if resp.StatusCode == http.StatusConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	// DBPilotからのレスポンスをパース
	var dbPilotResponse map[string]interface{}
	if err := json.Unmarshal(body, &dbPilotResponse); err != nil {
		logger.Logger.Error("レスポンスの解析に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process response"})
		return
	}

	logger.Logger.Info("アカウント作成が完了しました",
		append(logFields, zap.Any("response", dbPilotResponse))...)

	c.JSON(http.StatusOK, gin.H{
		"message": "Account created successfully",
		"user":    dbPilotResponse,
	})
}
