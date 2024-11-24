package handlers

import (
	"net/http"

	"dbpilot/logger"
	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UpdateUserRequest struct {
	Name     string `json:"name,omitempty"`     // 省略可能
	Password string `json:"password,omitempty"` // 省略可能
}

func UpdateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Logger.Warn("不正なユーザー更新リクエスト",
				zap.Error(err),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// セッション情報から更新対象のユーザーを特定
		sessionID, exists := c.Get("session")
		if !exists {
			logger.Logger.Warn("セッションが見つかりません",
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found"})
			return
		}

		// セッションIDからユーザー情報を取得
		var session models.LoginSession
		if err := db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
			logger.Logger.Error("セッション検証に失敗",
				zap.Error(err),
				zap.Any("session_id", sessionID),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			return
		}

		logger.Logger.Info("ユーザー更新リクエストを受信",
			zap.Uint("user_id", session.UserID),
			zap.Bool("password_update", req.Password != ""),
			zap.Bool("name_update", req.Name != ""),
		)

		// トランザクションを開始
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				logger.Logger.Error("パニックが発生したためロールバック",
					zap.Any("recover", r),
					zap.Uint("user_id", session.UserID),
				)
			}
		}()

		// パスワードの更新（存在する場合）
		if req.Password != "" {
			if err := tx.Model(&models.User{}).
				Where("id = ?", session.UserID).
				Update("password", req.Password).Error; err != nil {
				tx.Rollback()
				logger.Logger.Error("パスワード更新に失敗",
					zap.Error(err),
					zap.Uint("user_id", session.UserID),
				)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
				return
			}
			logger.Logger.Info("パスワードを更新しました",
				zap.Uint("user_id", session.UserID),
			)
		}

		// 名前の更新（存在する場合）
		if req.Name != "" {
			var profile models.Profile
			err := tx.Where("user_id = ?", session.UserID).First(&profile).Error

			if err == gorm.ErrRecordNotFound {
				// プロフィールが存在しない場合は新規作成
				profile = models.Profile{
					UserID: session.UserID,
					Name:   req.Name,
				}
				if err := tx.Create(&profile).Error; err != nil {
					tx.Rollback()
					logger.Logger.Error("プロフィール作成に失敗",
						zap.Error(err),
						zap.Uint("user_id", session.UserID),
					)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
					return
				}
				logger.Logger.Info("新規プロフィールを作成しました",
					zap.Uint("user_id", session.UserID),
					zap.String("name", req.Name),
				)
			} else if err != nil {
				tx.Rollback()
				logger.Logger.Error("プロフィール取得でエラーが発生",
					zap.Error(err),
					zap.Uint("user_id", session.UserID),
				)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			} else {
				// 既存のプロフィールを更新
				if err := tx.Model(&profile).Update("name", req.Name).Error; err != nil {
					tx.Rollback()
					logger.Logger.Error("プロフィール更新に失敗",
						zap.Error(err),
						zap.Uint("user_id", session.UserID),
					)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update name"})
					return
				}
				logger.Logger.Info("プロフィールを更新しました",
					zap.Uint("user_id", session.UserID),
					zap.String("name", req.Name),
				)
			}
		}

		// トランザクションをコミット
		if err := tx.Commit().Error; err != nil {
			logger.Logger.Error("トランザクションのコミットに失敗",
				zap.Error(err),
				zap.Uint("user_id", session.UserID),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		logger.Logger.Info("ユーザー情報の更新が完了しました",
			zap.Uint("user_id", session.UserID),
			zap.Bool("password_updated", req.Password != ""),
			zap.Bool("name_updated", req.Name != ""),
		)

		c.JSON(http.StatusOK, gin.H{
			"message": "User information updated successfully",
		})
	}
}
