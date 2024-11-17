package logger

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormZapLogger はZapロガーを使用するGORMのカスタムロガー
type GormZapLogger struct {
	ZapLogger *zap.Logger
	Config    gormlogger.Config
}

// NewGormZapLogger は新しいGormZapLoggerを作成します
func NewGormZapLogger(zapLogger *zap.Logger, config gormlogger.Config) gormlogger.Interface {
	return &GormZapLogger{
		ZapLogger: zapLogger,
		Config:    config,
	}
}

// LogMode はログモードを設定します
func (l *GormZapLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.Config.LogLevel = level
	return &newLogger
}

// Info は情報レベルのログを出力します
func (l GormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.Config.LogLevel >= gormlogger.Info {
		l.ZapLogger.Sugar().Infof(msg, data...)
	}
}

// Warn は警告レベルのログを出力します
func (l GormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.Config.LogLevel >= gormlogger.Warn {
		l.ZapLogger.Sugar().Warnf(msg, data...)
	}
}

// Error はエラーレベルのログを出力します
func (l GormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.Config.LogLevel >= gormlogger.Error {
		l.ZapLogger.Sugar().Errorf(msg, data...)
	}
}

// Trace はSQLクエリのログを出力します
func (l GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.Config.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// ログフィールドの準備
	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", elapsed),
	}

	// エラーがある場合はフィールドに追加
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fields = append(fields, zap.Error(err))
	}

	// スロークエリの判定
	if l.Config.SlowThreshold != 0 && elapsed > l.Config.SlowThreshold {
		l.ZapLogger.Warn("スロークエリを検出しました", fields...)
		return
	}

	// エラーの有無に応じたログレベルの選択
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.ZapLogger.Error("SQLクエリでエラーが発生しました", fields...)
	} else {
		l.ZapLogger.Debug("SQLクエリを実行しました", fields...)
	}
}
