package main

import (
	"context"
	"mailconvertor/config"
	"mailconvertor/handlers"
	"mailconvertor/logger"
	"mailconvertor/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	_, err := config.InitConfig()
	if err != nil {
		logger.Logger.Fatal("設定の初期化に失敗しました", zap.Error(err))
	}

	// ルーターの設定
	r := gin.New()
	r.Use(gin.Logger())

	// middleware 設定
	middlewareConfig := &middleware.Config{
		EnableLogger: true,
		EnableAuth:   true,
	}
	middleware.SetupMiddleware(r, middlewareConfig)

	r.POST("/receive", handlers.HandleEmailReceive)

	// サーバーの設定と起動
	srv := config.SetupServer(r)

	// グレースフルシャットダウンの実装
	handleGracefulShutdown(srv)
}

func handleGracefulShutdown(srv *http.Server) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// グレースフルシャットダウンの実行
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Error("サーバーのシャットダウンでエラーが発生", zap.Error(err))
	}

	logger.Logger.Info("サーバーを正常に終了しました")
}
