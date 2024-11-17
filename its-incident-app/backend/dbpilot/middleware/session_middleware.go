package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// セッション有効性を確認し、必要に応じてSERVICE_TOKENを検証するミドルウェア
func VerifySession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// AuthorizationヘッダーからセッションIDを取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		// "Bearer " プレフィックスを確認してセッションIDを抽出
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}
		sessionID := parts[1]

		// セッション情報をデータベースから取得
		var session models.LoginSession
		if err := db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
			// データベースにセッションがない場合、SERVICE_TOKENと比較
			serviceToken := os.Getenv("SERVICE_TOKEN")
			fmt.Print(serviceToken)
			if serviceToken == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Service token not configured"})
				c.Abort()
				return
			}
			if sessionID != serviceToken {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
				c.Abort()
				return
			}
			// SERVICE_TOKENと一致した場合は次のハンドラへ
			c.Set("session", &sessionID)
			c.Next()
			return
		}

		// 有効期限確認
		if time.Now().After(session.ExpiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
			c.Abort()
			return
		}

		// セッション情報をコンテキストに保存
		c.Set("session", &sessionID)

		// 次のハンドラへ
		c.Next()
	}
}
