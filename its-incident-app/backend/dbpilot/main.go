package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dbpilot/config"
	"dbpilot/handlers"
	"dbpilot/middleware"
	"dbpilot/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func setupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// 公開エンドポイント
	public := r.Group("/api/v1")
	{
		public.POST("/users", handlers.SaveUser(db))
		public.POST("/sessions", handlers.CreateSession(db))
		public.POST("/login", handlers.QueryUser(db))
		public.POST("/incidents", handlers.CreateIncident(db))
		public.POST("/emails", handlers.AddEmailHandler(db))

		// 新しい処理状態管理用エンドポイント
		public.GET("/status/:messageID", handlers.GetProcessingStatus(db))
		public.PUT("/status/:messageID", handlers.UpdateProcessingStatus(db))
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

		// レスポンス関連
		protected.POST("/responses", handlers.CreateResponse(db))

		// ユーザー関連
		protected.POST("/users-update", handlers.UpdateUser(db))
		protected.POST("/logout", handlers.LogoutHandler(db))

		// セッション関連
		protected.GET("/sessions", handlers.GetSession(db))
		protected.DELETE("/sessions", handlers.DeleteSession(db))
	}

	return r
}

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// データベース接続
	if err := config.ConnectDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// データベースのクリーンアップを保証
	defer func() {
		if err := config.CloseDatabase(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	db, err := config.GetDB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// マイグレーション
	if err := db.AutoMigrate(
		&models.User{},
		&models.Profile{},
		&models.LoginSession{},
		&models.Incident{},
		&models.Response{},
		&models.IncidentRelation{},
		&models.APIResponseData{},
		&models.ErrorLog{},
		&models.EmailData{},
		&models.ProcessingStatus{},
	); err != nil {
		log.Fatalf("Failed to perform database migration: %v", err)
	}

	// ルーターのセットアップ
	r := setupRouter(db)

	// サーバーの設定
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "3002"
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", serverPort),
		Handler: r,
	}

	// グレースフルシャットダウンの実装
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server is running on port %s", serverPort)

	// シグナルの待ち受け
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// シャットダウンのタイムアウト設定
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited properly")
}
