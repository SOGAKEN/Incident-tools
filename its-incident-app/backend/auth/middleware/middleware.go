package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"auth/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Config struct {
	EnableLogger bool
	EnableAuth   bool
	// 他のミドルウェア設定を追加
}

// SetupMiddleware ミドルウェアの設定
func SetupMiddleware(r *gin.Engine, cfg *Config) {
	// 基本的なミドルウェア
	r.Use(gin.Recovery())

	if cfg.EnableLogger {
		r.Use(GinLogger())
	}

	if cfg.EnableAuth {
		r.Use(AuthMiddleware())
	}
}

// AuthMiddleware Bearerトークン検証用ミドルウェア
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceToken := os.Getenv("SERVICE_TOKEN")
		if serviceToken == "" {
			logger.Logger.Warn("SERVICE_TOKEN is not set")
			abortWithError(c, http.StatusUnauthorized, "unauthorized")
			return
		}

		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logUnauthorizedRequest(c)
			abortWithError(c, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != serviceToken {
			logUnauthorizedRequest(c)
			abortWithError(c, http.StatusUnauthorized, "invalid token")
			return
		}

		c.Next()
	}
}

// abortWithError エラーレスポンスを返す補助関数
func abortWithError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}

// logUnauthorizedRequest 未認証リクエストのログ出力
func logUnauthorizedRequest(c *gin.Context) {
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	requestInfo := buildRequestInfo(c, bodyBytes)
	jsonData, err := json.MarshalIndent(requestInfo, "", "  ")
	if err != nil {
		logger.Logger.Error("リクエスト情報のJSON変換に失敗", zap.Error(err))
		return
	}

	logger.Logger.Warn("未認証リクエスト",
		zap.String("request_info", string(jsonData)),
		zap.String("client_ip", c.ClientIP()),
	)
}

// buildRequestInfo リクエスト情報の構築
type RequestInfo struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body,omitempty"`
}

func buildRequestInfo(c *gin.Context, bodyBytes []byte) RequestInfo {
	headers := make(map[string][]string)
	for name, values := range c.Request.Header {
		// センシティブなヘッダーの除外
		if !isProtectedHeader(name) {
			headers[name] = values
		}
	}

	return RequestInfo{
		Method:  c.Request.Method,
		Path:    c.Request.URL.Path,
		Headers: headers,
		Body:    string(bodyBytes),
	}
}

// isProtectedHeader センシティブなヘッダーかどうかを判定
func isProtectedHeader(header string) bool {
	sensitiveHeaders := map[string]bool{
		"Authorization": true,
		"Cookie":        true,
		// 他のセンシティブなヘッダーを追加
	}
	return sensitiveHeaders[header]
}

// GinLogger ロギングミドルウェア
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

		// トレース情報の追加
		if traceID := getTraceID(c); traceID != "" {
			fields = append(fields, zap.String("logging.googleapis.com/trace", traceID))
		}

		logRequestWithLevel(c, fields...)
	}
}

// getTraceID トレースIDの取得と整形
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

// logRequestWithLevel ステータスコードに応じたログレベルでログを出力
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

// SkipAuthMiddleware 特定のパスの認証をスキップするミドルウェアを生成
func SkipAuthMiddleware(skipPaths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 現在のパスがスキップ対象かチェック
		path := c.Request.URL.Path
		for _, skipPath := range skipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// AuthMiddlewareと同じ処理を実行
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Logger.Warn("認証ヘッダーが見つかりません")
			abortWithError(c, http.StatusUnauthorized, "unauthorized")
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			logUnauthorizedRequest(c)
			abortWithError(c, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != os.Getenv("SERVICE_TOKEN") {
			logUnauthorizedRequest(c)
			abortWithError(c, http.StatusUnauthorized, "invalid token")
			return
		}

		c.Next()
	}
}
