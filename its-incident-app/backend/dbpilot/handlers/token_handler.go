package handlers

import (
	"dbpilot/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResponseWrapper はレスポンスの共通構造体
type ResponseWrapper struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// handleError はエラーレスポンスを統一的に処理
func handleError(c *gin.Context, statusCode int, err error, additionalFields ...zap.Field) {
	fields := append([]zap.Field{
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Error(err),
	}, additionalFields...)

	logger.Logger.Error("エラーが発生しました", fields...)

	c.JSON(statusCode, ResponseWrapper{
		Success: false,
		Error:   err.Error(),
	})
}

// handleSuccess は成功レスポンスを統一的に処理
func handleSuccess(c *gin.Context, data interface{}, additionalFields ...zap.Field) {
	fields := append([]zap.Field{
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	}, additionalFields...)

	logger.Logger.Info("処理が成功しました", fields...)

	c.JSON(200, ResponseWrapper{
		Success: true,
		Data:    data,
	})
}
