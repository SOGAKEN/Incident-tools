package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"dbpilot/logger"
	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Config struct {
	EnableLogger  bool
	EnableSession bool // AuthMiddlewareを削除し、EnableSessionのみに
	DB            *gorm.DB
}

func SetupMiddleware(r *gin.Engine, cfg *Config) {
	r.Use(gin.Recovery())

	if cfg.EnableLogger {
		r.Use(GinLogger())
	}

	//	if cfg.EnableSession {
	//		r.Use(VerifySession(cfg.DB))
	//	}
}

// GinLogger はリクエストログを出力するミドルウェア
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", time.Since(start)),
			zap.String("user-agent", c.Request.UserAgent()),
		}

		if errors := c.Errors.ByType(gin.ErrorTypePrivate).String(); errors != "" {
			fields = append(fields, zap.String("errors", errors))
		}

		switch {
		case c.Writer.Status() >= 500:
			logger.Logger.Error("サーバーエラー", fields...)
		case c.Writer.Status() >= 400:
			logger.Logger.Warn("クライアントエラー", fields...)
		default:
			logger.Logger.Info("リクエスト完了", fields...)
		}
	}
}

// VerifySession はセッション検証を行うミドルウェア
func VerifySession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// AuthorizationヘッダーからセッションIDを取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Logger.Warn("認証ヘッダーが見つかりませんでした")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "認証ヘッダーが必要です"})
			c.Abort()
			return
		}

		// "Bearer " プレフィックスを確認してセッションIDを抽出
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			logger.Logger.Warn("認証ヘッダーの形式が正しくありません")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "認証ヘッダーの形式が正しくありません"})
			c.Abort()
			return
		}
		sessionID := parts[1]

		// セッション情報をデータベースから取得
		var session models.LoginSession
		if err := db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
			// データベースにセッションがない場合、SERVICE_TOKENと比較
			serviceToken := os.Getenv("SERVICE_TOKEN")
			if serviceToken == "" {
				logger.Logger.Error("サービストークンが設定されていません")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "サービストークンが設定されていません"})
				c.Abort()
				return
			}
			if sessionID != serviceToken {
				logger.Logger.Warn("セッションが無効です")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "セッションが無効です"})
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
			logger.Logger.Warn("セッションの有効期限が切れています")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "セッションの有効期限が切れています"})
			c.Abort()
			return
		}

		// セッション情報をコンテキストに保存
		c.Set("session", &sessionID)

		// 次のハンドラへ
		c.Next()
	}
}
