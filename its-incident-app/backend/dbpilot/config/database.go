package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDatabase はデータベースへの接続を確立します
func ConnectDatabase() error {
	// 必要な環境変数の検証
	requiredEnvVars := []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_PORT"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			return fmt.Errorf("required environment variable %s is not set", envVar)
		}
	}

	// ログレベルの設定
	logLevel := logger.Silent // デフォルトは Silent
	if os.Getenv("DEBUG") == "true" {
		logLevel = logger.Info
	}

	// カスタムロガーの設定
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // スロークエリとみなす閾値
			LogLevel:                  logLevel,    // ログレベル
			IgnoreRecordNotFoundError: true,        // record not found エラーを無視
			Colorful:                  false,       // カラー表示を無効化
		},
	)

	// データベース接続文字列の構築
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// GORMの設定
	config := &gorm.Config{
		Logger: newLogger,
		NowFunc: func() time.Time {
			jst, _ := time.LoadLocation("Asia/Tokyo")
			return time.Now().In(jst)
		},
	}

	// データベースへの接続
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 接続プールの設定
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// 接続プールの設定
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 接続テスト
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// GetDB は現在のデータベース接続を返します
func GetDB() (*gorm.DB, error) {
	if DB == nil {
		return nil, errors.New("database connection is not established")
	}
	return DB, nil
}

// CloseDatabase はデータベース接続を適切にクローズします
func CloseDatabase() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}
