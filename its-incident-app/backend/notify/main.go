package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification/config"
	"notification/handlers"
	"notification/logger"
	"notification/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 設定の初期化
	cfg, err := config.InitConfig()
	if err != nil {
		logger.Logger.Fatal("設定の初期化に失敗しました", zap.Error(err))
	}

	// ルーターの設定
	r := gin.New()
	r.Use(gin.Logger())

	// ミドルウェア設定
	middlewareConfig := &middleware.Config{
		EnableLogger: true,
		EnableAuth:   cfg.Environment == "production",
	}
	middleware.SetupMiddleware(r, middlewareConfig)

	// ハンドラーの設定
	r.POST("/send-login-link", handlers.SendLoginLink)
	r.POST("/notify", handlers.NotifyHandler)
	r.GET("/health", handleHealthCheck)

	// サーバーの設定と起動
	srv := config.SetupServer(r)

	// グレースフルシャットダウンの実装
	handleGracefulShutdown(srv, cfg.ShutdownTimeout)
}

// handleHealthCheck はヘルスチェックエンドポイントを処理します
func handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleGracefulShutdown(srv *http.Server, timeout time.Duration) {
	// サーバーを別のゴルーチンで起動
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("サーバーの起動に失敗しました", zap.Error(err))
		}
	}()

	// シグナルの受信設定
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Logger.Info("シャットダウンを開始します...")

	// シャットダウンのタイムアウト設定
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// グレースフルシャットダウンの実行
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Error("サーバーのシャットダウンでエラーが発生", zap.Error(err))
	}

	logger.Logger.Info("サーバーを正常に終了しました")
}
