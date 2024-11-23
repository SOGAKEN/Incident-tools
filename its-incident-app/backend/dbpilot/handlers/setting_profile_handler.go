// handlers/account_handler.go
package handlers

import (
	"dbpilot/logger"
	"dbpilot/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CreateAccountRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func CreateAccount(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "CreateAccount"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		var req CreateAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Logger.Error("リクエストのバリデーションに失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request format",
			})
			return
		}

		logFields = append(logFields,
			zap.String("email", req.Email),
			zap.String("name", req.Name))

		err := db.Transaction(func(tx *gorm.DB) error {
			// ユーザーの検索
			var user models.User
			if err := tx.Where("email = ?", req.Email).First(&user).Error; err != nil {
				logger.Logger.Error("ユーザーが見つかりません",
					append(logFields, zap.Error(err))...)
				return err
			}

			// パスワードの更新
			if err := tx.Model(&user).Update("password", req.Password).Error; err != nil {
				logger.Logger.Error("パスワードの更新に失敗しました",
					append(logFields, zap.Error(err))...)
				return err
			}

			// プロフィールの作成
			profile := models.Profile{
				UserID: user.ID,
				Name:   req.Name,
			}

			// プロフィールの作成または更新
			if err := tx.Where("user_id = ?", user.ID).
				Assign(models.Profile{Name: req.Name}).
				FirstOrCreate(&profile).Error; err != nil {
				logger.Logger.Error("プロフィールの作成/更新に失敗しました",
					append(logFields, zap.Error(err))...)
				return err
			}

			return nil
		})

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "User not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update account",
			})
			return
		}

		logger.Logger.Info("アカウント情報の更新が完了しました", logFields...)
		c.JSON(http.StatusOK, gin.H{
			"message": "Account updated successfully",
		})
	}
}
