package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// HandleAuthVerification DBPilotのセッション検証が成功した場合のレスポンスを返す
func (h *AuthHandler) HandleAuthVerification(c *gin.Context) {
	// DBPilotのミドルウェアでセットされたセッションIDを取得
	sessionID, exists := c.Get("session")
	if !exists {
		// セッションIDが存在しない場合はエラー（通常はDBPilotミドルウェアで弾かれているはず）
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Invalid session",
		})
		return
	}

	// 認証成功レスポンスを返す
	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"session_id": sessionID,
		"message":    "Authentication successful",
	})
}
