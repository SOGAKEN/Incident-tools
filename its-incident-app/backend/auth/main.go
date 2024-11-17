package main

import (
	"log"
	"os"

	"auth/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 環境変数のロード
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	r := gin.Default()

	// エンドポイント設定
	r.POST("/register", handlers.RegisterUser)
	r.POST("/login", handlers.LoginUser)
	r.POST("/update-user", handlers.UpdateUser)
	// r.POST("/logout", handlers.LogoutUser)
	r.GET("/verify-session", handlers.VerifySession)

	// サーバー起動
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080" // デフォルトポート
	}
	r.Run(":" + serverPort)
}
