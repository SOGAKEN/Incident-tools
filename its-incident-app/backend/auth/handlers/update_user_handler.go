// auth-service/handlers/update_user_handler.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// クライアントからのリクエスト構造体
type UpdateUserRequest struct {
	NewName         string `json:"name,omitempty"`     // 新しい名前（オプション）
	NewPassword     string `json:"password,omitempty"` // 新しいパスワード（オプション）
	CurrentPassword string `json:"current_password"`   // 現在のパスワード（必須）
}

// DB Pilotへのリクエスト構造体
type DBPilotUpdateRequest struct {
	Name     string `json:"name,omitempty"`     // 更新する名前
	Password string `json:"password,omitempty"` // ハッシュ化されたパスワード
}

func UpdateUser(c *gin.Context) {
	var userReq UpdateUserRequest
	if err := c.ShouldBindJSON(&userReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// リクエストの検証
	if userReq.NewName == "" && userReq.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No update data provided"})
		return
	}

	if userReq.CurrentPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is required"})
		return
	}

	// Auth ServiceからDB Pilotへのリクエストには、クライアントから受け取ったセッションIDをそのまま使用
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	baseURL := os.Getenv("DB_PILOT_SERVICE_URL")
	if baseURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Pilot Service URL is not configured"})
		return
	}

	// DB Pilot Serviceへの更新リクエストを準備
	updateReq := DBPilotUpdateRequest{}

	// パスワードの更新がある場合
	if userReq.NewPassword != "" {
		// パスワードをハッシュ化
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userReq.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		updateReq.Password = string(hashedPassword)
	}

	// 名前の更新がある場合
	if userReq.NewName != "" {
		updateReq.Name = userReq.NewName
	}

	// DB Pilotへリクエストを送信
	updateReqJSON, _ := json.Marshal(updateReq)
	httpReq, err := http.NewRequest(http.MethodPost, baseURL+"/users-update", bytes.NewBuffer(updateReqJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request to DB Pilot"})
		return
	}

	// DB PilotへセッションIDを転送
	httpReq.Header.Set("Authorization", authHeader)
	httpReq.Header.Set("Content-Type", "application/json")

	// リクエストを実行
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to communicate with DB Pilot"})
		return
	}
	defer resp.Body.Close()

	// DB Pilotからのレスポンスを確認
	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			c.JSON(resp.StatusCode, gin.H{"error": "Update failed"})
			return
		}
		c.JSON(resp.StatusCode, gin.H{"error": errorResponse.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User information updated successfully",
	})
}