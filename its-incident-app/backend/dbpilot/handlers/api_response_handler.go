package handlers

import (
	"dbpilot/logger"
	"dbpilot/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func GetAPIResponseData(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "GetAPIResponseData"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		var query models.APIResponseDataQuery
		if err := c.ShouldBindJSON(&query); err != nil {
			logger.Logger.Error("リクエストのバインドに失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// クエリの構築
		dbQuery := db.Model(&models.APIResponseData{})

		// ID関連の検索条件
		if query.IncidentID != nil {
			dbQuery = dbQuery.Where("incident_id = ?", *query.IncidentID)
		}
		if query.TaskID != nil {
			dbQuery = dbQuery.Where("task_id = ?", *query.TaskID)
		}
		if query.WorkflowRunID != nil {
			dbQuery = dbQuery.Where("workflow_run_id = ?", *query.WorkflowRunID)
		}
		if query.WorkflowID != nil {
			dbQuery = dbQuery.Where("workflow_id = ?", *query.WorkflowID)
		}
		if query.Status != nil {
			dbQuery = dbQuery.Where("status = ?", *query.Status)
		}
		if query.IncidentNumber != nil {
			dbQuery = dbQuery.Where("incident_number = ?", *query.IncidentNumber)
		}

		// テキストフィールドの検索（ILIKE使用）
		textFields := map[string]*string{
			"body":          query.Body,
			"user":          query.User,
			"host":          query.Host,
			"priority":      query.Priority,
			"subject":       query.Subject,
			"from":          query.From,
			"place":         query.Place,
			"incident_text": query.IncidentText,
			"time":          query.Time,
			"judgment":      query.Judgment,
			"sender":        query.Sender,
			"final":         query.Final,
		}

		for field, value := range textFields {
			if value != nil && *value != "" {
				dbQuery = dbQuery.Where(field+" ILIKE ?", "%"+*value+"%")
			}
		}

		// 数値範囲の検索
		if query.ElapsedTimeMin != nil {
			dbQuery = dbQuery.Where("elapsed_time >= ?", *query.ElapsedTimeMin)
		}
		if query.ElapsedTimeMax != nil {
			dbQuery = dbQuery.Where("elapsed_time <= ?", *query.ElapsedTimeMax)
		}
		if query.TotalTokensMin != nil {
			dbQuery = dbQuery.Where("total_tokens >= ?", *query.TotalTokensMin)
		}
		if query.TotalTokensMax != nil {
			dbQuery = dbQuery.Where("total_tokens <= ?", *query.TotalTokensMax)
		}
		if query.TotalStepsMin != nil {
			dbQuery = dbQuery.Where("total_steps >= ?", *query.TotalStepsMin)
		}
		if query.TotalStepsMax != nil {
			dbQuery = dbQuery.Where("total_steps <= ?", *query.TotalStepsMax)
		}

		// 時間範囲の検索
		if query.CreatedAtStart != nil {
			dbQuery = dbQuery.Where("created_at >= ?", *query.CreatedAtStart)
		}
		if query.CreatedAtEnd != nil {
			dbQuery = dbQuery.Where("created_at <= ?", *query.CreatedAtEnd)
		}
		if query.FinishedAtStart != nil {
			dbQuery = dbQuery.Where("finished_at >= ?", *query.FinishedAtStart)
		}
		if query.FinishedAtEnd != nil {
			dbQuery = dbQuery.Where("finished_at <= ?", *query.FinishedAtEnd)
		}

		// ソート
		if query.SortBy != nil && *query.SortBy != "" {
			direction := "ASC"
			if query.SortDirection != nil && *query.SortDirection == "desc" {
				direction = "DESC"
			}
			dbQuery = dbQuery.Order(*query.SortBy + " " + direction)
		} else {
			// デフォルトのソート
			dbQuery = dbQuery.Order("created_at DESC")
		}

		// ページネーション
		limit := 100 // デフォルト
		if query.Limit != nil && *query.Limit > 0 {
			limit = *query.Limit
		}
		if query.Offset != nil {
			dbQuery = dbQuery.Offset(*query.Offset)
		}
		dbQuery = dbQuery.Limit(limit)

		// 総件数の取得
		var total int64
		if err := dbQuery.Count(&total).Error; err != nil {
			logger.Logger.Error("総件数の取得に失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total count"})
			return
		}

		// データの取得
		var apiResponses []models.APIResponseData
		if err := dbQuery.Find(&apiResponses).Error; err != nil {
			logger.Logger.Error("APIレスポンスデータの取得に失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch API response data"})
			return
		}

		logger.Logger.Info("APIレスポンスデータを取得しました",
			append(logFields,
				zap.Int("count", len(apiResponses)),
				zap.Int64("total", total))...)

		c.JSON(http.StatusOK, gin.H{
			"total":  total,
			"count":  len(apiResponses),
			"offset": query.Offset,
			"limit":  limit,
			"data":   apiResponses,
		})
	}
}
