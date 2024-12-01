package handlers

import (
	"dbpilot/logger"
	"dbpilot/models"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    string `json:"code,omitempty"`
}

// エラーハンドリング用のヘルパー関数
func logAndReturnError(c *gin.Context, statusCode int, err error, code string, logFields []zap.Field) {
	logger.Logger.Error("エラーが発生しました",
		append(logFields,
			zap.Error(err),
			zap.String("error_code", code))...)

	c.JSON(statusCode, ErrorResponse{
		Error: err.Error(),
		Code:  code,
	})
}

// トランザクション処理用のヘルパー関数
func withTransaction(db *gorm.DB, c *gin.Context, logFields []zap.Field, fn func(*gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		logAndReturnError(c, http.StatusInternalServerError, tx.Error, "DB_TRANSACTION_ERROR", logFields)
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Logger.Error("パニックが発生しました",
				append(logFields, zap.Any("recover", r))...)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		logAndReturnError(c, http.StatusInternalServerError, err, "COMMIT_ERROR", logFields)
		return err
	}

	return nil
}

// 単一インシデント取得ハンドラー
func GetIncident(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "GetIncident"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		// パスパラメータからIDを取得
		idStr := c.Param("id")

		// IDをログフィールドに追加
		logFields = append(logFields, zap.String("incident_id", idStr))

		var incident models.Incident
		err := db.Preload("Responses").
			Preload("Relations").
			Preload("Relations.RelatedIncident").
			Preload("APIData").
			First(&incident, "message_id = ?", idStr).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Logger.Info("インシデントが見つかりませんでした", logFields...)
				c.JSON(http.StatusNotFound, gin.H{"error": "インシデントが見つかりません"})
			} else {
				logAndReturnError(c, http.StatusInternalServerError, err, "FETCH_ERROR", logFields)
			}
			return
		}

		logger.Logger.Info("インシデントを取得しました",
			append(logFields,
				zap.String("status", incident.Status),
				zap.String("assignee", incident.Assignee))...)

		c.JSON(http.StatusOK, incident)
	}
}

// インシデント一覧取得ハンドラー
func GetIncidentAll(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "GetIncidentAll"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		var req struct {
			Page      int      `json:"page"`
			Limit     int      `json:"limit"`
			Status    []string `json:"status"`
			From      string   `json:"from"`
			To        string   `json:"to"`
			Assignees []string `json:"assignee"` // 複数のassigneeを配列で受け取る
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			logAndReturnError(c, http.StatusBadRequest, err, "INVALID_REQUEST", logFields)
			return
		}

		// 検索条件のログ
		logFields = append(logFields,
			zap.Int("page", req.Page),
			zap.Int("limit", req.Limit),
			zap.Strings("status", req.Status),
			zap.Strings("assignee", req.Assignees)) // assigneesのログ

		// ページネーション設定
		if req.Page < 1 {
			req.Page = 1
		}
		if req.Limit <= 0 {
			req.Limit = 10
		}
		offset := (req.Page - 1) * req.Limit

		// 日付処理
		fromTime, toTime, err := parseDateRange(req.From, req.To, logFields)
		if err != nil {
			logAndReturnError(c, http.StatusBadRequest, err, "INVALID_DATE", logFields)
			return
		}

		var (
			incidents    []models.Incident
			total        int64
			statusCounts []struct {
				Status string `json:"status"`
				Count  int64  `json:"count"`
			}
			uniqueAssignees []string
		)

		// トランザクション処理
		err = withTransaction(db, c, logFields, func(tx *gorm.DB) error {
			// 有効なインシデントIDを取得
			validIncidentIDs := tx.Model(&models.APIResponseData{}).
				Select("incident_id").
				Where("subject IS NOT NULL AND subject != ''")

			// メインクエリ構築
			query := tx.Model(&models.Incident{}).
				Where("id IN (?)", validIncidentIDs)

			// 検索条件の追加
			if len(req.Status) > 0 {
				query = query.Where("status IN (?)", req.Status)
			}
			if !fromTime.IsZero() || !toTime.Equal(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)) {
				query = query.Where("datetime BETWEEN ? AND ?", fromTime, toTime)
			}
			// 複数のassigneeによる検索
			if len(req.Assignees) > 0 {
				query = query.Where("assignee IN (?)", req.Assignees)
			}

			// 総数取得
			if err := query.Count(&total).Error; err != nil {
				return err
			}

			// ステータスカウント取得のクエリを構築
			statusCountQuery := tx.Model(&models.Incident{}).
				Where("id IN (?)", validIncidentIDs)

			// assigneesフィルターがある場合はステータスカウントにも適用
			if len(req.Assignees) > 0 {
				statusCountQuery = statusCountQuery.Where("assignee IN (?)", req.Assignees)
			}

			if err := statusCountQuery.
				Select("status, count(*) as count").
				Group("status").
				Scan(&statusCounts).Error; err != nil {
				return err
			}

			// assigneeのユニーク値を取得
			if err := tx.Model(&models.Incident{}).
				Where("assignee IS NOT NULL AND assignee != ''"). // NULLや空文字を除外
				Distinct("assignee").
				Order("assignee ASC"). // assigneeでソート
				Pluck("assignee", &uniqueAssignees).Error; err != nil {
				return err
			}

			// データ取得
			return query.Preload("Responses").
				Preload("Relations").
				Preload("Relations.RelatedIncident").
				Preload("APIData").
				Order("id DESC").
				Limit(req.Limit).
				Offset(offset).
				Find(&incidents).Error
		})

		if err != nil {
			return // エラーは既にレスポンス済み
		}

		logger.Logger.Info("インシデント一覧を取得しました",
			append(logFields,
				zap.Int64("total", total),
				zap.Int("count", len(incidents)))...)

		c.Header("Cache-Control", "private, max-age=300")
		c.JSON(http.StatusOK, gin.H{
			"data": incidents,
			"meta": gin.H{
				"total": total,
				"page":  req.Page,
				"limit": req.Limit,
				"pages": (total + int64(req.Limit) - 1) / int64(req.Limit),
			},
			"status_counts":    statusCounts,
			"unique_assignees": uniqueAssignees,
		})
	}
}

// 日付範囲パース用のヘルパー関数
func parseDateRange(fromStr, toStr string, logFields []zap.Field) (time.Time, time.Time, error) {
	var fromTime, toTime time.Time
	layout := "2006-01-02 15:04"

	if strings.TrimSpace(fromStr) != "" {
		var err error
		fromTime, err = time.Parse(layout, strings.TrimSpace(fromStr))
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid 'from' date format: %v", err)
		}
	}

	if strings.TrimSpace(toStr) != "" {
		var err error
		toTime, err = time.Parse(layout, strings.TrimSpace(toStr))
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid 'to' date format: %v", err)
		}
	} else {
		toTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	if !fromTime.IsZero() && !toTime.IsZero() && fromTime.After(toTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date range: 'from' date is after 'to' date")
	}

	return fromTime, toTime, nil
}
