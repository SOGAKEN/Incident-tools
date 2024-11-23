package config

import (
	"dbpilot/logger"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ServerConfig struct {
	Port            string
	GinMode         string
	LogLevel        zapcore.Level
	Environment     string
	ProjectID       string
	ServiceName     string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

// InitConfig は環境設定を初期化します
func InitConfig() (*ServerConfig, error) {
	// .envファイルの読み込み
	if err := godotenv.Load(); err != nil {
		fmt.Println(".envファイルが見つかりません")
	}

	// データベース接続
	if err := ConnectDatabase(); err != nil {
		return nil, fmt.Errorf("データベース接続に失敗: %w", err)
	}

	// ログレベルの設定
	logLevel := initLogLevel()

	// Ginモードの設定
	ginMode := initGinMode()

	return &ServerConfig{
		Port:            getEnv("SERVER_PORT", "8080"),
		GinMode:         ginMode,
		LogLevel:        logLevel,
		Environment:     getEnv("ENVIRONMENT", "development"),
		ServiceName:     getEnv("K_SERVICE", "dbpilot"),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		ReadTimeout:     getDuration("HTTP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    getDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:     getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
	}, nil
}

// SetupServer はサーバーの設定を行います
func SetupServer(r *gin.Engine) *http.Server {
	config, err := InitConfig()
	if err != nil {
		logger.Logger.Fatal("サーバー設定の初期化に失敗しました", zap.Error(err))
	}

	// ルート情報の表示
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
		"=================================\n"+
		"%s"+
		"=================================\n",
		config.Port,
		config.GinMode,
		logger.LogLevel.String(),
		config.Environment,
		config.ServiceName,
		routeInfo.String())
}
