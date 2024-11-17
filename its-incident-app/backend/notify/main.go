package main

import (
	"log"
	"notify/handlers"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 環境変数のロード
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.Default()

	// Notificationエンドポイントの設定
	r.POST("/notify", handlers.NotifyHandler)

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080" // デフォルトポート
	}
	r.Run(":" + serverPort)
}
