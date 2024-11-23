package handlers

import (
	"auth/logger"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AddAccountRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type DBPilotRequest struct {
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type NotificationRequest struct {
	Email     string `json:"email"`
	Token     string `json:"token"`
	LoginURL  string `json:"login_url"`
	ExpiresIn string `json:"expires_in"`
}

// トークン生成関数
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func AddAccountUser(c *gin.Context) {
	logFields := []zap.Field{
		zap.String("handler", "AddAccountUser"),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	}

	// Bearerトークンの取得
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		logger.Logger.Error("認証ヘッダーが見つかりません", logFields...)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	var req AddAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Error("リクエストのバインドに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// トークンの生成
	token, err := generateToken()
	if err != nil {
		logger.Logger.Error("トークン生成に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// DB Pilotへ送信するデータの準備
	expiresAt := time.Now().Add(60 * time.Minute)
	dbReqBody := DBPilotRequest{
		Email:     req.Email,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	// DB PilotへのリクエストJSONの作成
	jsonData, err := json.Marshal(dbReqBody)
	if err != nil {
		logger.Logger.Error("JSONエンコードに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare DB Pilot request"})
		return
	}

	// DB Pilotへのリクエスト作成
	dbPilotURL := os.Getenv("DB_PILOT_SERVICE_URL") + "/login-tokens"
	dbReq, err := http.NewRequest("POST", dbPilotURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Logger.Error("DB Pilotリクエストの作成に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create DB Pilot request"})
		return
	}

	// ヘッダーの設定
	dbReq.Header.Set("Content-Type", "application/json")
	dbReq.Header.Set("Authorization", authHeader)

	// DB Pilotへリクエスト送信
	client := &http.Client{}
	resp, err := client.Do(dbReq)
	if err != nil {
		logger.Logger.Error("DB Pilotへのリクエスト送信に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to DB Pilot"})
		return
	}
	defer resp.Body.Close()

	// DB Pilotからのレスポンスチェック
	if resp.StatusCode != http.StatusOK {
		logger.Logger.Error("DB Pilotからエラーレスポンスを受信しました",
			append(logFields, zap.Int("status_code", resp.StatusCode))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token in DB Pilot"})
		return
	}

	// Notification Serviceへのリクエスト準備
	loginURL := fmt.Sprintf("%s/auth/verify?token=%s",
		os.Getenv("FRONTEND_URL"), token)

	notifReqBody := NotificationRequest{
		Email:     req.Email,
		Token:     token,
		LoginURL:  loginURL,
		ExpiresIn: "15分",
	}

	notificationJSON, err := json.Marshal(notifReqBody)
	if err != nil {
		logger.Logger.Error("通知リクエストのJSONエンコードに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare notification request"})
		return
	}

	// Notification Serviceへのリクエスト作成
	notificationURL := os.Getenv("NOTIFICATION_SERVICE_URL") + "/send-login-link"
	notifReq, err := http.NewRequest("POST", notificationURL, bytes.NewBuffer(notificationJSON))
	if err != nil {
		logger.Logger.Error("通知サービスリクエストの作成に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification request"})
		return
	}

	// 通知サービスへのヘッダー設定
	notifReq.Header.Set("Content-Type", "application/json")
	notifReq.Header.Set("Authorization", authHeader)

	// 通知サービスへリクエスト送信
	notificationResp, err := client.Do(notifReq)
	if err != nil {
		logger.Logger.Error("通知サービスへのリクエスト送信に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}
	defer notificationResp.Body.Close()

	// 通知サービスからのレスポンスチェック
	if notificationResp.StatusCode != http.StatusOK {
		logger.Logger.Error("通知サービスからエラーレスポンスを受信しました",
			append(logFields, zap.Int("status_code", notificationResp.StatusCode))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send login email"})
		return
	}

	logger.Logger.Info("ログインリンクの送信を完了しました",
		append(logFields, zap.String("email", req.Email))...)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login link has been sent to your email",
	})
}
