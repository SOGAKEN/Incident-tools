package config

import (
	"fmt"
	"mailconvertor/logger"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

	// ログレベルの設定
	logLevel := initLogLevel()

	// Ginモードの設定
	ginMode := initGinMode()

	return &ServerConfig{
		Port:        getServerPort(),
		GinMode:     ginMode,
		LogLevel:    logLevel,
		Environment: getEnv("ENVIRONMENT", "development"),
		ServiceName: getEnv("K_SERVICE", "mailconvertor"),
	}, nil
}

// SetupServer はサーバーの設定を行います
func SetupServer(r *gin.Engine) *http.Server {
	config, _ := InitConfig()
	displayServerConfig(r, config)

	return &http.Server{
		Addr:              ":" + config.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func initLogLevel() zapcore.Level {
	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "info"
	}

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

func getServerPort() string {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return "8080"
	}
	return serverPort
}
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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
		"- Environment: %s\n"+ // 追加
		"- Service: %s\n"+ // 追加
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
