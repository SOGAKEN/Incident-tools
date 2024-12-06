package config

import (
	"autopilot/logger"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap/zapcore"
)

// AIConfig はAIサービスの設定を管理します
// AI処理の再試行とタイムアウトの設定を含みます
type AIConfig struct {
	// サービスの接続設定
	Endpoint string
	Token    string

	// タイムアウト設定
	ShortTimeout time.Duration // 短時間の処理用タイムアウト
	LongTimeout  time.Duration // 長時間の処理用タイムアウト

	// リトライ設定
	MaxRetries    int           // 最大再試行回数
	MinRetryDelay time.Duration // 最小再試行待機時間
	MaxRetryDelay time.Duration // 最大再試行待機時間
}

// ServerConfig はアプリケーション全体の設定を管理します
type ServerConfig struct {
	// サーバー基本設定
	Port        string
	GinMode     string
	LogLevel    zapcore.Level
	Environment string
	ProjectID   string
	ServiceName string

	// 外部サービス接続設定
	DBPilotURL   string
	ServiceToken string

	// タイムアウト設定
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration

	// AIサービス設定
	AI AIConfig
}

// InitConfig は環境設定を初期化します
// 環境変数から設定を読み込み、デフォルト値との組み合わせで設定を構築します
func InitConfig() (*ServerConfig, error) {
	// .envファイルの読み込み（存在する場合）
	if err := godotenv.Load(); err != nil {
		fmt.Println(".envファイルが見つかりません")
	}

	// 基本設定の初期化
	logLevel := initLogLevel()
	ginMode := initGinMode()

	// 設定オブジェクトの構築
	config := &ServerConfig{
		// サーバー基本設定
		Port:        getEnv("SERVER_PORT", "8080"),
		GinMode:     ginMode,
		LogLevel:    logLevel,
		Environment: getEnv("ENVIRONMENT", "development"),
		ProjectID:   getEnv("GOOGLE_CLOUD_PROJECT", ""),
		ServiceName: getEnv("K_SERVICE", "auto-service"),

		// 外部サービス接続設定
		DBPilotURL:   getEnv("DBPILOT_URL", ""),
		ServiceToken: getEnv("SERVICE_TOKEN", ""),

		// タイムアウト設定
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		ReadTimeout:     getDuration("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    getDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:     getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),

		// AIサービス設定
		AI: AIConfig{
			// 接続設定
			Endpoint: getEnv("AI_ENDPOINT", ""),
			Token:    getEnv("AI_TOKEN", ""),

			// タイムアウト設定
			ShortTimeout: getDuration("AI_SHORT_TIMEOUT", 30*time.Second),
			LongTimeout:  getDuration("AI_LONG_TIMEOUT", 90*time.Second),

			// リトライ設定
			MaxRetries:    getIntEnv("AI_MAX_RETRIES", 3),
			MinRetryDelay: getDuration("AI_MIN_RETRY_DELAY", 2*time.Second),
			MaxRetryDelay: getDuration("AI_MAX_RETRY_DELAY", 5*time.Second),
		},
	}

	// 設定の検証と返却
	return config, config.Validate()
}

// SetupServer はHTTPサーバーを設定します
func SetupServer(r *gin.Engine) *http.Server {
	config, _ := InitConfig()
	displayServerConfig(r, config)

	return &http.Server{
		Addr:              ":" + config.Port,
		Handler:           r,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

// Validate は設定値の妥当性を検証します
func (c *ServerConfig) Validate() error {
	// 必須項目の検証
	required := map[string]string{
		"DBPilotURL":   c.DBPilotURL,
		"ServiceToken": c.ServiceToken,
		"AIEndpoint":   c.AI.Endpoint,
		"AIToken":      c.AI.Token,
	}

	for name, value := range required {
		if value == "" {
			return fmt.Errorf("%s is required", name)
		}
	}

	// AIサービス設定の妥当性検証
	if err := c.validateAIConfig(); err != nil {
		return fmt.Errorf("invalid AI configuration: %v", err)
	}

	return nil
}

// validateAIConfig はAI設定の妥当性を検証します
func (c *ServerConfig) validateAIConfig() error {
	if c.AI.MaxRetries < 1 {
		return fmt.Errorf("AI_MAX_RETRIES must be greater than 0")
	}

	if c.AI.MinRetryDelay >= c.AI.MaxRetryDelay {
		return fmt.Errorf("AI_MIN_RETRY_DELAY must be less than AI_MAX_RETRY_DELAY")
	}

	if c.AI.ShortTimeout >= c.AI.LongTimeout {
		return fmt.Errorf("AI_SHORT_TIMEOUT must be less than AI_LONG_TIMEOUT")
	}

	return nil
}

// getIntEnv は環境変数から整数値を取得します
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		fmt.Printf("Invalid value for %s: '%s', using default: %d\n",
			key, value, defaultValue)
	}
	return defaultValue
}

// displayServerConfig は現在の設定を表示します
func displayServerConfig(r *gin.Engine, config *ServerConfig) {
	var routeInfo strings.Builder
	routeInfo.WriteString("Registered Endpoints:\n")
	for _, route := range r.Routes() {
		routeInfo.WriteString(fmt.Sprintf("- %s: %s -> %s\n",
			route.Method,
			route.Path,
			route.Handler))
	}

	fmt.Printf("\n"+
		"=================================\n"+
		"Server Configuration:\n"+
		"- Port: %s\n"+
		"- Mode: %s\n"+
		"- Log Level: %s\n"+
		"- Environment: %s\n"+
		"- Service: %s\n"+
		"\nAI Configuration:\n"+
		"- Max Retries: %d\n"+
		"- Retry Delay: %v-%v\n"+
		"- Timeouts: short=%v, long=%v\n"+
		"=================================\n"+
		"%s"+
		"=================================\n",
		config.Port,
		config.GinMode,
		logger.LogLevel.String(),
		config.Environment,
		config.ServiceName,
		config.AI.MaxRetries,
		config.AI.MinRetryDelay,
		config.AI.MaxRetryDelay,
		config.AI.ShortTimeout,
		config.AI.LongTimeout,
		routeInfo.String())
}

// 既存のヘルパー関数
func initLogLevel() zapcore.Level {
	logLevelStr := getEnv("LOG_LEVEL", "info")
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(logLevelStr)); err != nil {
		fmt.Printf("Invalid LOG_LEVEL '%s', defaulting to 'info'\n", logLevelStr)
		logLevel = zapcore.InfoLevel
	}
	logger.LogLevel.SetLevel(logLevel)
	return logLevel
}

func initGinMode() string {
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "release"
	}
	gin.SetMode(ginMode)
	return ginMode
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
