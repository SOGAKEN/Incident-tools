// logger/logger.go

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// ログレベルを保持する変数
	LogLevel = zap.NewAtomicLevel()
	// Loggerはグローバルなロガーです
	Logger *zap.Logger
)

func init() {
	// Zapの設定を作成
	config := zap.NewProductionConfig()

	// ログレベルを設定
	config.Level = LogLevel

	// 出力をstdoutに設定（Cloud Runはstdoutからログを収集）
	config.OutputPaths = []string{"stdout"}

	// Encoderの設定（Cloud Loggingのフォーマットに合わせる）
	config.EncoderConfig = zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "severity",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // INFO, WARN, ERRORなど
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// ロガーを構築
	var err error
	Logger, err = config.Build()
	if err != nil {
		panic(err)
	}

	// グローバルロガーを置き換え
	zap.ReplaceGlobals(Logger)
}
