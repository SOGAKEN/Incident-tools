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

// DBPilotResponse はDBPilotからのレスポンスを格納する構造体
type DBPilotResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Email    string `json:"email"`
		UserID   uint   `json:"user_id"`
		Name     string `json:"name,omitempty"`
		ImageURL string `json:"image_url,omitempty"`
	} `json:"data"`
	Error string `json:"error,omitempty"`
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

	// レスポンスボディの読み取り
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Logger.Error("レスポンスの読み取りに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response",
		})
		return
	}

	logFields = append(logFields, zap.String("response_body", string(respBody)))

	// レスポンスのステータスコードチェック
	if resp.StatusCode != http.StatusOK {
		logger.Logger.Error("トークン検証に失敗しました",
			append(logFields,
				zap.Int("status_code", resp.StatusCode))...)

		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(respBody, &errorResponse); err == nil && errorResponse.Error != "" {
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
	var dbpilotResponse DBPilotResponse
	if err := json.Unmarshal(respBody, &dbpilotResponse); err != nil {
		logger.Logger.Error("レスポンスのデコードに失敗しました",
			append(logFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process response",
		})
		return
	}

	// データのチェック
	if !dbpilotResponse.Success {
		logger.Logger.Error("トークン検証が失敗しました", logFields...)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": dbpilotResponse.Error,
		})
		return
	}

	if dbpilotResponse.Data.Email == "" {
		logger.Logger.Error("メールアドレスが取得できませんでした", logFields...)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Email not found in response",
		})
		return
	}

	logger.Logger.Info("トークンの検証が成功しました",
		append(logFields,
			zap.String("email", dbpilotResponse.Data.Email),
			zap.Uint("user_id", dbpilotResponse.Data.UserID))...)

	// 成功レスポンスの返却
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"email":     dbpilotResponse.Data.Email,
			"user_id":   dbpilotResponse.Data.UserID,
			"name":      dbpilotResponse.Data.Name,
			"image_url": dbpilotResponse.Data.ImageURL,
		},
	})
}
