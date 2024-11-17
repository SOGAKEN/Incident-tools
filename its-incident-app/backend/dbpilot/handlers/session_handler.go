// db-pilot-service/handlers/session_handler.go
package handlers

import (
	"net/http"
	"time"

	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateSessionRequest struct {
	UserID    uint      `json:"user_id" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
	SessionID string    `json:"session_id" binding:"required"`
	ExpiresAt time.Time `json:"expires_at" binding:"required"`
}

// CreateSession は新しいセッションをDBに保存します
func CreateSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		// セッション情報を構造体に格納
		session := &models.LoginSession{
			UserID:    req.UserID,
			Email:     req.Email,
			SessionID: req.SessionID,
			ExpiresAt: req.ExpiresAt,
		}

		// モデルの CreateSession メソッドを使用して保存
		if err := models.CreateSession(db, session); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create session",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Session created successfully",
			"session": gin.H{
				"id":         session.ID,
				"created_at": session.CreatedAt,
				"user_id":    session.UserID,
				"email":      session.Email,
				"session_id": session.SessionID,
				"expires_at": session.ExpiresAt,
			},
		})
	}
}

// GetSession はセッション情報を取得します
func GetSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Query("email")
		if email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			return
		}

		session, err := models.GetSessionByEmail(db, email)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
			return
		}

		c.JSON(http.StatusOK, session)
	}
}

// DeleteSession はセッションを削除します
func DeleteSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Query("email")
		if email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			return
		}

		if err := models.DeleteSessionByEmail(db, email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Session deleted successfully"})
	}
}