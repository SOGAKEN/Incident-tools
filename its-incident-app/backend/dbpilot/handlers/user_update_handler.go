package handlers

import (
	"net/http"

	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UpdateUserRequest はユーザー情報更新のリクエスト構造体
type UpdateUserRequest struct {
	Name     string `json:"name,omitempty"`     // 省略可能
	Password string `json:"password,omitempty"` // 省略可能
}

func UpdateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// セッション情報から更新対象のユーザーを特定
		sessionID, exists := c.Get("session")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found"})
			return
		}

		// セッションIDからユーザー情報を取得
		var session models.LoginSession
		if err := db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			return
		}

		// トランザクションを開始
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// パスワードの更新（存在する場合）
		if req.Password != "" {
			if err := tx.Model(&models.User{}).
				Where("id = ?", session.UserID).
				Update("password", req.Password).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
				return
			}
		}

		// 名前の更新（存在する場合）
		if req.Name != "" {
			// プロフィールが存在するか確認
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
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
					return
				}
			} else if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			} else {
				// 既存のプロフィールを更新
				if err := tx.Model(&profile).Update("name", req.Name).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update name"})
					return
				}
			}
		}

		// トランザクションをコミット
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User information updated successfully",
		})
	}
}