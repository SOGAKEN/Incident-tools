package handlers

import (
	"net/http"
	"time"

	"dbpilot/logger"
	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
			logger.Logger.Warn("不正なセッション作成リクエスト",
				zap.Error(err),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		logger.Logger.Info("セッション作成リクエストを受信",
			zap.Uint("user_id", req.UserID),
			zap.String("email", req.Email),
			zap.Time("expires_at", req.ExpiresAt),
		)

		// セッション情報を構造体に格納
		session := &models.LoginSession{
			UserID:    req.UserID,
			Email:     req.Email,
			SessionID: req.SessionID,
			ExpiresAt: req.ExpiresAt,
		}

		// モデルの CreateSession メソッドを使用して保存
		if err := models.CreateSession(db, session); err != nil {
			logger.Logger.Error("セッション作成に失敗",
				zap.Error(err),
				zap.Uint("user_id", req.UserID),
				zap.String("email", req.Email),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create session",
				"details": err.Error(),
			})
			return
		}

		logger.Logger.Info("セッションを作成しました",
			zap.Uint("session_db_id", session.ID),
			zap.String("session_id", session.SessionID),
			zap.Uint("user_id", session.UserID),
			zap.Time("expires_at", session.ExpiresAt),
		)

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
			logger.Logger.Warn("メールアドレスが指定されていません",
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			return
		}

		logger.Logger.Info("セッション取得リクエストを受信",
			zap.String("email", email),
		)

		session, err := models.GetSessionByEmail(db, email)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				logger.Logger.Warn("セッションが見つかりません",
					zap.String("email", email),
				)
				c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
				return
			}
			logger.Logger.Error("セッション取得に失敗",
				zap.Error(err),
				zap.String("email", email),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
			return
		}

		logger.Logger.Info("セッションを取得しました",
			zap.String("email", email),
			zap.String("session_id", session.SessionID),
			zap.Time("expires_at", session.ExpiresAt),
		)

		c.JSON(http.StatusOK, session)
	}
}

// DeleteSession はセッションを削除します
func DeleteSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Query("email")
		if email == "" {
			logger.Logger.Warn("メールアドレスが指定されていません",
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			return
		}

		logger.Logger.Info("セッション削除リクエストを受信",
			zap.String("email", email),
		)

		if err := models.DeleteSessionByEmail(db, email); err != nil {
			logger.Logger.Error("セッション削除に失敗",
				zap.Error(err),
				zap.String("email", email),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
			return
		}

		logger.Logger.Info("セッションを削除しました",
			zap.String("email", email),
		)

		c.JSON(http.StatusOK, gin.H{"message": "Session deleted successfully"})
	}
}
