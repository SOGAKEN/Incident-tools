// middleware/middleware.go

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

	"auth/config"
	"auth/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Config struct {
	EnableLogger bool
	EnableAuth   bool
	ServerConfig *config.ServerConfig
}

// SetupMiddleware ミドルウェアの設定
func SetupMiddleware(r *gin.Engine, cfg *Config) {
	r.Use(gin.Recovery())

	if cfg.EnableLogger {
		r.Use(GinLogger())
	}

	if cfg.EnableAuth {
		r.Use(AuthMiddleware(cfg.ServerConfig))
	}
}

// authenticateRequest リクエストの認証を行う共通関数
func authenticateRequest(c *gin.Context, cfg *config.ServerConfig) bool {
	// SERVICE_TOKENによる認証チェック
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		serviceToken := os.Getenv("SERVICE_TOKEN")
		if serviceToken != "" && token == serviceToken {
			return true
		}
	}

	// セッションクッキーによる認証チェック
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		if err := verifySessionWithDBPilot(c, cfg.DBPilotURL, sessionID); err == nil {
			c.Set("session", sessionID)
			return true
		} else {
			logger.Logger.Warn("セッション検証に失敗",
				zap.Error(err),
				zap.String("session_id", sessionID),
			)
		}
	}

	return false
}

// AuthMiddleware 認証ミドルウェア
func AuthMiddleware(cfg *config.ServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authenticateRequest(c, cfg) {
			c.Next()
			return
		}

		logUnauthorizedRequest(c)
		abortWithError(c, http.StatusUnauthorized, "unauthorized")
	}
}

// SkipAuthMiddleware 認証スキップミドルウェア
func SkipAuthMiddleware(cfg *config.ServerConfig, skipPaths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 現在のパスがスキップ対象かチェック
		path := c.Request.URL.Path
		for _, skipPath := range skipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		if authenticateRequest(c, cfg) {
			c.Next()
			return
		}

		logUnauthorizedRequest(c)
		abortWithError(c, http.StatusUnauthorized, "unauthorized")
	}
}

// verifySessionWithDBPilot セッション検証
func verifySessionWithDBPilot(c *gin.Context, dbpilotURL, sessionID string) error {
	if dbpilotURL == "" {
		return fmt.Errorf("DB_PILOT_SERVICE_URL is not configured")
	}

	verifyURL := fmt.Sprintf("%s/api/v1/sessions/verify", dbpilotURL)
	req, err := http.NewRequest("GET", verifyURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create verification request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sessionID))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send verification request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("session verification failed with status %d: %s",
			resp.StatusCode, string(body))
	}

	return nil
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

// RequestInfo リクエスト情報の構造体
type RequestInfo struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body,omitempty"`
}

// buildRequestInfo リクエスト情報の構築
func buildRequestInfo(c *gin.Context, bodyBytes []byte) RequestInfo {
	headers := make(map[string][]string)
	for name, values := range c.Request.Header {
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
		"Set-Cookie":    true,
	}
	return sensitiveHeaders[strings.ToLower(header)]
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
