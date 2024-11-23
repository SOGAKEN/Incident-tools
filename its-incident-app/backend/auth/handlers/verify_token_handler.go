package handlers

import (
	"auth/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TokenVerificationResponse struct {
	Email  string `json:"email"`
	UserID uint   `json:"user_id,omitempty"`
}

func VerifyToken(c *gin.Context) {
	logFields := []zap.Field{
		zap.String("handler", "VerifyToken"),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	}

	// トークンの取得
	token := c.Query("token")
	if token == "" {
		logger.Logger.Error("トークンが指定されていません", logFields...)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token is required",
		})
		return
	}

	logFields = append(logFields, zap.String("token", token))

	// DBPilotに検証リクエストを送信
	dbPilotURL := fmt.Sprintf("%s/login-tokens/verify?token=%s",
		os.Getenv("DB_PILOT_SERVICE_URL"), token)

	// DBPilotへのリクエスト作成
	req, err := http.NewRequest("GET", dbPilotURL, nil)
	if err != nil {
		logger.Logger.Error("リクエストの作成に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create request",
		})
		return
	}

	// ヘッダーの設定
	req.Header.Set("Content-Type", "application/json")

	// リクエスト送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.Error("DB Pilotへのリクエスト送信に失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify token",
		})
		return
	}
	defer resp.Body.Close()

	// レスポンスボディの読み取り（エラーメッセージのために）
	respBody, _ := io.ReadAll(resp.Body)
	logFields = append(logFields, zap.String("response_body", string(respBody)))

	// レスポンスのステータスコードチェック
	if resp.StatusCode != http.StatusOK {
		logger.Logger.Error("トークン検証に失敗しました",
			append(logFields,
				zap.Int("status_code", resp.StatusCode))...)

		// DBPilotからのエラーメッセージを解析
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(respBody, &errorResponse); err == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errorResponse.Error,
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
		}
		return
	}

	// レスポンスのデコード
	var verificationResponse TokenVerificationResponse
	if err := json.Unmarshal(respBody, &verificationResponse); err != nil {
		logger.Logger.Error("レスポンスのデコードに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process response",
		})
		return
	}

	if verificationResponse.Email == "" {
		logger.Logger.Error("メールアドレスが取得できませんでした", logFields...)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Email not found in response",
		})
		return
	}

	logger.Logger.Info("トークンの検証が成功しました",
		append(logFields, zap.String("email", verificationResponse.Email))...)

	c.JSON(http.StatusOK, gin.H{
		"message": "Token verified successfully",
		"email":   verificationResponse.Email,
		"user_id": verificationResponse.UserID,
	})
}
