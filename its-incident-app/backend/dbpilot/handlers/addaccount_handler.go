package handlers

import (
	"dbpilot/models"
	"fmt"
	"net/http"
	"time"

	"dbpilot/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CreateLoginToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			handleError(c, http.StatusBadRequest, err)
			return
		}

		// トランザクションを開始
		err := db.Transaction(func(tx *gorm.DB) error {
			// ユーザーを検索または作成
			var user models.User
			if err := tx.FirstOrCreate(&user, models.User{
				Email: req.Email,
			}).Error; err != nil {
				return err
			}

			// 既存の有効期限切れトークンを無効化
			if err := tx.Model(&models.LoginToken{}).
				Where("email = ? AND expires_at <= ? AND is_expired = ?",
					req.Email, time.Now(), false).
				Update("is_expired", true).Error; err != nil {
				return err
			}

			// 新しいトークンを作成
			loginToken := &models.LoginToken{
				Email:     req.Email,
				Token:     req.Token,
				ExpiresAt: req.ExpiresAt,
				IsExpired: false,
			}

			if err := tx.Create(loginToken).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			handleError(c, http.StatusInternalServerError, err)
			return
		}

		handleSuccess(c, gin.H{
			"message": "Login token created successfully",
			"email":   req.Email,
		})
	}
}

func VerifyLoginToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "VerifyLoginToken"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		// キャッシュ制御ヘッダーの設定
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		token := c.Query("token")
		if token == "" {
			logger.Logger.Error("トークンが指定されていません", logFields...)
			c.JSON(http.StatusBadRequest, ResponseWrapper{
				Success: false,
				Error:   "Token is required",
			})
			return
		}

		clientIP := c.ClientIP()
		logFields = append(logFields,
			zap.String("token", token),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", c.Request.UserAgent()))

		var user models.User
		var loginToken models.LoginToken
		var userProfile models.Profile

		err := db.Transaction(func(tx *gorm.DB) error {
			// トークンを取得
			if err := tx.Where("token = ?", token).First(&loginToken).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return fmt.Errorf("invalid token")
				}
				return err
			}

			logFields = append(logFields,
				zap.Time("expires_at", loginToken.ExpiresAt),
				zap.Bool("is_expired", loginToken.IsExpired),
				zap.String("email", loginToken.Email))

			logger.Logger.Info("トークンの状態を確認しました", logFields...)

			// 有効期限チェック
			if time.Now().After(loginToken.ExpiresAt) {
				if err := tx.Model(&loginToken).Update("is_expired", true).Error; err != nil {
					return fmt.Errorf("failed to update token expiration: %w", err)
				}
				return fmt.Errorf("token expired")
			}

			if loginToken.IsExpired {
				return fmt.Errorf("token expired")
			}

			// ユーザー情報とプロフィールを取得
			if err := tx.Where("email = ?", loginToken.Email).First(&user).Error; err != nil {
				return fmt.Errorf("user not found: %w", err)
			}

			// プロフィール情報を取得（存在しない場合はスキップ）
			tx.Where("user_id = ?", user.ID).First(&userProfile)

			// アクセスログを記録
			tokenAccess := &models.TokenAccess{
				TokenID:    loginToken.ID,
				Token:      loginToken.Token,
				Email:      loginToken.Email,
				IP:         clientIP,
				UserAgent:  c.Request.UserAgent(),
				AccessedAt: time.Now(),
			}

			if err := tx.Create(tokenAccess).Error; err != nil {
				logger.Logger.Warn("アクセスログの記録に失敗しました",
					append(logFields, zap.Error(err))...)
				// アクセスログの記録失敗は処理を継続
			}

			return nil
		})

		if err != nil {
			switch err.Error() {
			case "invalid token":
				logger.Logger.Error("無効なトークンです", logFields...)
				c.JSON(http.StatusUnauthorized, ResponseWrapper{
					Success: false,
					Error:   "Invalid token",
				})
			case "token expired":
				logger.Logger.Error("トークンの有効期限が切れています", logFields...)
				c.JSON(http.StatusUnauthorized, ResponseWrapper{
					Success: false,
					Error:   "Token has expired",
				})
			default:
				logger.Logger.Error("予期せぬエラーが発生しました",
					append(logFields, zap.Error(err))...)
				c.JSON(http.StatusInternalServerError, ResponseWrapper{
					Success: false,
					Error:   "Internal server error",
				})
			}
			return
		}

		logger.Logger.Info("トークンの検証が成功しました",
			append(logFields,
				zap.Uint("user_id", user.ID),
				zap.String("email", user.Email))...)

		// レスポンスの作成
		response := models.TokenVerificationResponse{
			Email:    user.Email,
			UserID:   user.ID,
			Name:     userProfile.Name,
			ImageURL: userProfile.ImageURL,
		}

		handleSuccess(c, response)
	}
}
