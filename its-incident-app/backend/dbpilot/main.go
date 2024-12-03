package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dbpilot/config"
	"dbpilot/handlers"
	"dbpilot/logger"
	"dbpilot/middleware"
	"dbpilot/migrations"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// 設定の初期化
	cfg, err := config.InitConfig()
	if err != nil {
		logger.Logger.Fatal("設定の初期化に失敗しました",
			zap.Error(err),
			zap.String("service", cfg.ServiceName),
			zap.String("environment", cfg.Environment),
		)
	}

	// ログレベルの設定
	if err := logger.LogLevel.UnmarshalText([]byte(cfg.LogLevel.String())); err != nil {
		logger.Logger.Fatal("ログレベルの設定に失敗しました",
			zap.Error(err),
			zap.String("intended_level", cfg.LogLevel.String()),
		)
	}

	// データベースの初期化
	db, err := config.GetDB()
	if err != nil {
		logger.Logger.Fatal("データベースの取得に失敗しました",
			zap.Error(err),
		)
	}

	// データベースのクリーンアップを保証
	defer func() {
		if err := config.CloseDatabase(); err != nil {
			logger.Logger.Error("データベース接続のクローズに失敗しました",
				zap.Error(err),
			)
		}
	}()

	// マイグレーション
	if err := performMigrations(db); err != nil {
		logger.Logger.Fatal("マイグレーションに失敗しました",
			zap.Error(err),
		)
	}

	// ルーターの設定
	r := setupRouter(db, cfg)

	// サーバーの設定と起動（config.SetupServerを使用）
	srv := config.SetupServer(r)

	// アプリケーション情報のログ出力
	logger.Logger.Info("アプリケーションを起動します",
		zap.String("service", cfg.ServiceName),
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
		zap.String("gin_mode", cfg.GinMode),
	)

	// グレースフルシャットダウンの実装
	handleGracefulShutdown(srv, cfg.ShutdownTimeout)
}

func setupRouter(db *gorm.DB, cfg *config.ServerConfig) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	// 基本的なミドルウェア設定
	middlewareConfig := &middleware.Config{
		EnableLogger: true,
		DB:           db,
	}
	middleware.SetupMiddleware(r, middlewareConfig)
	authHandler := handlers.NewAuthHandler()

	logger.Logger.Info("ルーターの設定を開始します")

	// 公開エンドポイント
	public := r.Group("/api/v1")
	{
		public.POST("/users", handlers.SaveUser(db))
		public.POST("/login", handlers.QueryUser(db))
		public.POST("/incidents", handlers.CreateIncident(db))
		public.POST("/emails", handlers.AddEmailHandler(db))
		public.GET("/status/:messageID", handlers.GetProcessingStatus(db))
		public.PUT("/status/:messageID", handlers.UpdateProcessingStatus(db))
		public.POST("/login-tokens", handlers.CreateLoginToken(db))
		public.GET("/login-tokens/verify", handlers.VerifyLoginToken(db))
		public.POST("/accounts", handlers.CreateAccount(db))
		public.POST("/sessions", handlers.CreateSession(db))
	}

	// 保護されたエンドポイント
	protected := r.Group("/api/v1")
	protected.Use(middleware.VerifySession(db))
	{
		// プロフィール関連
		protected.POST("/profiles", handlers.RegisterProfile(db))
		protected.GET("/profiles", handlers.GetProfile(db))

		// インシデント関連
		protected.GET("/incidents/:id", handlers.GetIncident(db))
		protected.POST("/incidents-all", handlers.GetIncidentAll(db))
		protected.POST("/incident-relations", handlers.CreateIncidentRelation(db))

		protected.POST("/email-all", handlers.GetEmailDataAll(db))
		protected.GET("/email/:message_id", handlers.GetEmailDataWithIncident(db))

		// レスポンス関連
		protected.POST("/responses", handlers.CreateResponse(db))

		// ユーザー関連
		protected.POST("/users-update", handlers.UpdateUser(db))
		protected.POST("/logout", handlers.LogoutHandler(db))

		// セッション関連
		protected.GET("/sessions", handlers.GetSession(db))
		protected.DELETE("/sessions", handlers.DeleteSession(db))

		// Workflows用のエンドポイント
		protected.POST("/api-responses/search", handlers.GetAPIResponseData(db))
		protected.GET("/sessions/verify", authHandler.HandleAuthVerification)

	}

	logger.Logger.Info("ルーターの設定が完了しました")
	return r
}

func performMigrations(db *gorm.DB) error {
	logger.Logger.Info("データベースマイグレーションを開始します")

	// マイグレーションマネージャーの作成
	migrator := migrations.NewMigrator(db)

	// マイグレーションの実行
	if err := migrator.RunMigrations(); err != nil {
		return err
	}

	logger.Logger.Info("データベースマイグレーションが完了しました")
	return nil
}

func handleGracefulShutdown(srv *http.Server, timeout time.Duration) {
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("サーバーの起動に失敗しました", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info("シャットダウンを開始します...", zap.Duration("timeout", timeout))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Error("サーバーのシャットダウンでエラーが発生", zap.Error(err))
	}

	logger.Logger.Info("サーバーを正常に終了しました")
}
