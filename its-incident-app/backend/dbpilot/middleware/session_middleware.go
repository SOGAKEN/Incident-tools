// middleware/middleware.go

package middleware

import (
	"bytes"
	"fmt"
	"io"
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
	EnableLogger bool
	DB           *gorm.DB
}

// SetupMiddleware はミドルウェアの基本設定を行います
func SetupMiddleware(r *gin.Engine, cfg *Config) {
	r.Use(gin.Recovery())

	if cfg.EnableLogger {
		r.Use(GinLogger())
	}
}

// GinLogger はリクエストログを出力するミドルウェア
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// リクエストボディの読み取りと復元
		var bodyBytes []byte
		if c.Request.Body != nil && shouldLogBody(path) {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		// 基本的なログフィールド
		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", time.Since(start)),
			zap.String("user-agent", c.Request.UserAgent()),
		}

		// Cloud Trace IDの追加
		if traceID := getTraceID(c); traceID != "" {
			fields = append(fields, zap.String("logging.googleapis.com/trace", traceID))
		}

		// エラー情報の追加
		if errors := c.Errors.ByType(gin.ErrorTypePrivate).String(); errors != "" {
			fields = append(fields, zap.String("errors", errors))
		}

		// ヘッダー情報の追加（センシティブ情報を除外）
		headers := filterHeaders(c.Request.Header)
		if len(headers) > 0 {
			fields = append(fields, zap.Any("headers", headers))
		}

		logRequestWithLevel(c, fields...)
	}
}

// VerifySession はセッション検証を行うミドルウェア
func VerifySession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logUnauthorizedRequest(c, "認証ヘッダーが見つかりませんでした")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "認証ヘッダーが必要です"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			logUnauthorizedRequest(c, "認証ヘッダーの形式が正しくありません")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "認証ヘッダーの形式が正しくありません"})
			c.Abort()
			return
		}
		sessionID := parts[1]

		// サービストークンチェック
		serviceToken := os.Getenv("SERVICE_TOKEN")
		if serviceToken != "" && sessionID == serviceToken {
			c.Set("session", sessionID) // セッションIDのみを保存
			c.Next()
			return
		}

		var session models.LoginSession
		if err := db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				logUnauthorizedRequest(c, "セッションが見つかりません")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			} else {
				logger.Logger.Error("セッション検証でエラーが発生しました",
					zap.Error(err),
					zap.String("session_id", sessionID),
				)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
			c.Abort()
			return
		}

		if time.Now().After(session.ExpiresAt) {
			logUnauthorizedRequest(c, "セッションの有効期限が切れています")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
			c.Abort()
			return
		}

		// セッションIDのみをコンテキストに保存
		c.Set("session", session.SessionID)
		c.Next()
	}
}

// logUnauthorizedRequest は未認証リクエストのログを出力します
func logUnauthorizedRequest(c *gin.Context, message string) {
	requestInfo := gin.H{
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
		"query":     c.Request.URL.RawQuery,
		"client_ip": c.ClientIP(),
		"message":   message,
	}

	logger.Logger.Warn("未認証リクエスト",
		zap.Any("request_info", requestInfo),
		zap.String("client_ip", c.ClientIP()),
	)
}

// Helper functions

func getTraceID(c *gin.Context) string {
	traceHeader := c.Request.Header.Get("X-Cloud-Trace-Context")
	if traceHeader == "" {
		return ""
	}

	traceParts := strings.Split(traceHeader, "/")
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" || len(traceParts) == 0 {
		return ""
	}

	return fmt.Sprintf("projects/%s/traces/%s", projectID, traceParts[0])
}

func filterHeaders(headers http.Header) map[string][]string {
	filtered := make(map[string][]string)
	for name, values := range headers {
		if !isProtectedHeader(name) {
			filtered[name] = values
		}
	}
	return filtered
}

func isProtectedHeader(header string) bool {
	sensitiveHeaders := map[string]bool{
		"Authorization": true,
		"Cookie":        true,
		"Set-Cookie":    true,
	}
	return sensitiveHeaders[strings.ToLower(header)]
}

func shouldLogBody(path string) bool {
	// ヘルスチェックなど、ボディのログが不要なパスを除外
	excludedPaths := map[string]bool{
		"/health": true,
		"/ping":   true,
	}
	return !excludedPaths[path]
}

func logRequestWithLevel(c *gin.Context, fields ...zap.Field) {
	switch {
	case c.Writer.Status() >= 500:
		logger.Logger.Error("サーバーエラー", fields...)
	case c.Writer.Status() >= 400:
		logger.Logger.Warn("クライアントエラー", fields...)
	default:
		logger.Logger.Info("リクエスト完了", fields...)
	}
}
