package handlers

import (
	"auth/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterUser(c *gin.Context) {
	logFields := []zap.Field{
		zap.String("handler", "RegisterUser"),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	}

	// 認証ヘッダーの取得
	authHeader := c.GetHeader("Authorization")
	// Bearer トークンの抽出
	var token string
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
		logFields = append(logFields, zap.String("bearer_token", token))
	}

	// リクエストのバインド
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Error("リクエストのバインドに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	logFields = append(logFields, zap.String("email", req.Email))
	logger.Logger.Info("ユーザー登録を開始します", logFields...)

	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Logger.Error("パスワードのハッシュ化に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password encryption failed"})
		return
	}

	// DB Pilot Serviceへのリクエスト準備
	baseURL := os.Getenv("DB_PILOT_SERVICE_URL")
	saveUserReq := map[string]string{
		"email":    req.Email,
		"password": string(hashedPassword),
	}

	saveUserReqJSON, err := json.Marshal(saveUserReq)
	if err != nil {
		logger.Logger.Error("リクエストのJSONエンコードに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare request"})
		return
	}

	// HTTPリクエストの作成
	request, err := http.NewRequest("POST", baseURL+"/users", bytes.NewBuffer(saveUserReqJSON))
	if err != nil {
		logger.Logger.Error("DBPilotへのリクエスト作成に失敗しました",
			append(logFields,
				zap.String("url", baseURL+"/users"),
				zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// ヘッダーの設定
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	// リクエストの実行
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Logger.Error("DBPilotへのリクエスト送信に失敗しました",
			append(logFields,
				zap.String("url", baseURL+"/users"),
				zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to DB Pilot Service"})
		return
	}
	defer resp.Body.Close()

	// レスポンスの確認
	if resp.StatusCode != http.StatusOK {
		// レスポンスボディの読み取り
		body, _ := io.ReadAll(resp.Body)
		logger.Logger.Error("DBPilotからエラーレスポンスを受信しました",
			append(logFields,
				zap.Int("status_code", resp.StatusCode),
				zap.String("response_body", string(body)))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user to DB Pilot Service"})
		return
	}

	logger.Logger.Info("ユーザー登録が完了しました", logFields...)
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}
