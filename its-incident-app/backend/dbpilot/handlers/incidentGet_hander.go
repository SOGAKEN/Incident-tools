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

const (
	ErrInvalidStatus  = "INVALID_STATUS"
	ErrStatusNotFound = "STATUS_NOT_FOUND"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    string `json:"code,omitempty"`
}

// logAndReturnError エラーログとレスポンスを処理するヘルパー関数
func logAndReturnError(c *gin.Context, statusCode int, err error, code string, logFields []zap.Field) {
	// システムエラーはERROR、クライアントエラーはWARNレベルでログ出力
	if statusCode >= 500 {
		logger.Logger.Error("システムエラーが発生しました",
			append(logFields,
				zap.Error(err),
				zap.String("error_code", code))...)
	} else {
		logger.Logger.Warn("クライアントエラーが発生しました",
			append(logFields,
				zap.Error(err),
				zap.String("error_code", code))...)
	}

	c.JSON(statusCode, ErrorResponse{
		Error: err.Error(),
		Code:  code,
	})
}

// withTransaction トランザクション処理用のヘルパー関数
func withTransaction(db *gorm.DB, c *gin.Context, logFields []zap.Field, fn func(*gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		logAndReturnError(c, http.StatusInternalServerError, tx.Error, "DB_TRANSACTION_ERROR", logFields)
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			// パニックはERRORレベル
			logger.Logger.Error("トランザクション中にパニックが発生しました",
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

	// トランザクション成功はDEBUGレベル
	logger.Logger.Debug("トランザクションが正常に完了しました", logFields...)
	return nil
}

// GetIncident 単一インシデント取得ハンドラー
func GetIncident(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "GetIncident"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		// リクエスト開始はDEBUGレベル
		logger.Logger.Debug("インシデント取得を開始します", logFields...)

		idStr := c.Param("id")
		logFields = append(logFields, zap.String("incident_id", idStr))

		var incident models.Incident
		err := db.Preload("Responses").
			Preload("Relations").
			Preload("Relations.RelatedIncident").
			Preload("APIData").
			Preload("Status").
			First(&incident, "message_id = ?", idStr).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logAndReturnError(c, http.StatusBadRequest, errors.New("指定されたステータスが存在しません"), ErrStatusNotFound, logFields)
				return
			}
			logAndReturnError(c, http.StatusInternalServerError, err, ErrInvalidStatus, logFields)
			return
		}

		// 正常取得はINFOレベル（重要な業務イベント）
		logger.Logger.Info("インシデントを取得しました",
			append(logFields,
				zap.String("status", incident.Status.Name),
				zap.String("assignee", incident.Assignee))...)

		c.JSON(http.StatusOK, incident)
	}
}

// GetIncidentAll インシデント一覧取得ハンドラー
func GetIncidentAll(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "GetIncidentAll"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		logger.Logger.Debug("インシデント一覧取得を開始します", logFields...)

		var req struct {
			Page      int      `json:"page"`
			Limit     int      `json:"limit"`
			Status    []string `json:"status"`
			From      string   `json:"from"`
			To        string   `json:"to"`
			Assignees []string `json:"assignee"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			logAndReturnError(c, http.StatusBadRequest, err, "INVALID_REQUEST", logFields)
			return
		}

		if req.Page < 1 {
			req.Page = 1
		}
		if req.Limit <= 0 {
			req.Limit = 10
		}
		offset := (req.Page - 1) * req.Limit

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

		err = withTransaction(db, c, logFields, func(tx *gorm.DB) error {
			validIncidentIDs := tx.Model(&models.APIResponseData{}).
				Select("incident_id").
				Where("subject IS NOT NULL AND subject != ''")

			query := tx.Model(&models.Incident{}).
				Where("id IN (?)", validIncidentIDs)

			if len(req.Status) > 0 {
				var statusIDs []uint
				if err := tx.Model(&models.IncidentStatus{}).
					Where("name IN ?", req.Status).
					Pluck("id", &statusIDs).Error; err != nil {
					return err
				}
				query = query.Where("status_id IN ?", statusIDs)
			}
			if !fromTime.IsZero() || !toTime.Equal(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)) {
				query = query.Where("datetime BETWEEN ? AND ?", fromTime, toTime)
			}
			if len(req.Assignees) > 0 {
				query = query.Where("assignee IN (?)", req.Assignees)
			}

			if err := query.Count(&total).Error; err != nil {
				return err
			}

			// ステータスごとの件数を取得
			statusCountQuery := tx.Table("incidents AS i").
				Joins("JOIN incident_statuses AS s ON i.status_id = s.id").
				Select("s.name AS status, COUNT(*) AS count").
				Where("i.id IN (?)", validIncidentIDs).
				Group("s.name")

			if err := statusCountQuery.Scan(&statusCounts).Error; err != nil {
				return err
			}

			if err := tx.Model(&models.Incident{}).
				Where("assignee IS NOT NULL AND assignee != ''").
				Distinct("assignee").
				Order("assignee ASC").
				Pluck("assignee", &uniqueAssignees).Error; err != nil {
				return err
			}

			return query.Preload("Responses").
				Preload("Relations").
				Preload("Relations.RelatedIncident").
				Preload("APIData").
				Preload("Status").
				Order("id DESC").
				Limit(req.Limit).
				Offset(offset).
				Find(&incidents).Error
		})

		if err != nil {
			return
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

// parseDateRange 日付範囲パース用のヘルパー関数
func parseDateRange(fromStr, toStr string, logFields []zap.Field) (time.Time, time.Time, error) {
	var fromTime, toTime time.Time
	layout := "2006-01-02 15:04"

	// パース処理の開始はDEBUGレベル
	logger.Logger.Debug("日付範囲のパースを開始します",
		append(logFields,
			zap.String("from", fromStr),
			zap.String("to", toStr))...)

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

// EmailDATA
func GetEmailDataAll(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "GetEmailDataAll"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		logger.Logger.Debug("EmailData一覧取得を開始します", logFields...)

		var req struct {
			Page      int      `json:"page"`
			Limit     int      `json:"limit"`
			Status    []string `json:"status"`
			From      string   `json:"from"`
			To        string   `json:"to"`
			Assignees []string `json:"assignee"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			logAndReturnError(c, http.StatusBadRequest, err, "INVALID_REQUEST", logFields)
			return
		}

		if req.Page < 1 {
			req.Page = 1
		}
		if req.Limit <= 0 {
			req.Limit = 10
		}
		offset := (req.Page - 1) * req.Limit

		fromTime, toTime, err := parseDateRange(req.From, req.To, logFields)
		if err != nil {
			logAndReturnError(c, http.StatusBadRequest, err, "INVALID_DATE", logFields)
			return
		}

		var (
			emailDataList []models.EmailData
			total         int64
			statusCounts  []struct {
				Status string `json:"status"`
				Count  int64  `json:"count"`
			}
			uniqueAssignees []string
		)

		err = withTransaction(db, c, logFields, func(tx *gorm.DB) error {
			query := tx.Model(&models.EmailData{}).
				Joins("LEFT JOIN incidents ON email_data.message_id = incidents.message_id").
				Joins("LEFT JOIN incident_statuses ON incidents.status_id = incident_statuses.id")

			// フィルタリング
			if len(req.Status) > 0 {
				query = query.Where("incident_statuses.name IN ?", req.Status)
			}
			if !fromTime.IsZero() || !toTime.IsZero() {
				query = query.Where("email_data.created_at BETWEEN ? AND ?", fromTime, toTime)
			}
			if len(req.Assignees) > 0 {
				query = query.Where("incidents.assignee IN ?", req.Assignees)
			}

			if err := query.Count(&total).Error; err != nil {
				return err
			}

			var currentStatusOrder int
			if len(req.Status) > 0 {
				err := tx.Model(&models.IncidentStatus{}).
					Where("name IN ?", req.Status).
					Select("MAX(display_order)").
					Scan(&currentStatusOrder).Error
				if err != nil {
					return err
				}
			}

			// ステータスごとの件数を取得（前のステータスのみ）
			statusCountQuery := tx.Table("incidents AS i").
				Joins("JOIN incident_statuses AS s ON i.status_id = s.id").
				Select("s.name AS status, COUNT(*) AS count").
				Group("s.name")

			if err := statusCountQuery.Scan(&statusCounts).Error; err != nil {
				return err
			}

			// 担当者の一覧を取得
			assigneeQuery := tx.Model(&models.Incident{}).
				Where("assignee IS NOT NULL AND assignee != ''").
				Distinct("assignee").
				Order("assignee ASC")

			if len(req.Status) > 0 {
				var statusIDs []uint
				if err := tx.Model(&models.IncidentStatus{}).
					Where("name IN ?", req.Status).
					Pluck("id", &statusIDs).Error; err != nil {
					return err
				}
				assigneeQuery = assigneeQuery.Where("status_id IN ?", statusIDs)
			}
			if !fromTime.IsZero() || !toTime.IsZero() {
				assigneeQuery = assigneeQuery.Where("datetime BETWEEN ? AND ?", fromTime, toTime)
			}

			if err := assigneeQuery.Pluck("assignee", &uniqueAssignees).Error; err != nil {
				return err
			}

			// データの取得
			err = query.Preload("Incident.Status").
				Preload("Incident.APIData").
				Offset(offset).
				Limit(req.Limit).
				Order("email_data.created_at DESC").
				Find(&emailDataList).Error

			return err
		})

		if err != nil {
			return
		}

		totalPages := (total + int64(req.Limit) - 1) / int64(req.Limit)

		logger.Logger.Info("EmailData一覧を取得しました",
			append(logFields,
				zap.Int64("total", total),
				zap.Int("count", len(emailDataList)))...)

		c.Header("Cache-Control", "private, max-age=300")
		c.JSON(http.StatusOK, gin.H{
			"data": emailDataList,
			"meta": gin.H{
				"total": total,
				"page":  req.Page,
				"limit": req.Limit,
				"pages": totalPages,
			},
			"status_counts":    statusCounts,
			"unique_assignees": uniqueAssignees,
		})
	}
}

// エラーコードの定義
const (
	ErrEmailDataNotFoundCode = "ErrEmailDataNotFound"
	ErrInvalidEmailDataCode  = "ErrInvalidEmailData"
)

// エラーメッセージの定義
var (
	ErrEmailDataNotFound = errors.New("EmailDataが見つかりません")
	ErrInvalidEmailData  = errors.New("無効なEmailDataです")
)

func GetEmailDataWithIncident(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "GetEmailDataWithIncident"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		// リクエスト開始はDEBUGレベル
		logger.Logger.Debug("EmailData取得を開始します", logFields...)

		messageID := c.Param("message_id")
		logFields = append(logFields, zap.String("message_id", messageID))

		var emailData models.EmailData

		err := db.Preload("Incident.Responses").
			Preload("Incident.Relations").
			Preload("Incident.Relations.RelatedIncident").
			Preload("Incident.APIData").
			Preload("Incident.Status").
			Where("message_id = ?", messageID).
			First(&emailData).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logAndReturnError(c, http.StatusNotFound, ErrEmailDataNotFound, ErrEmailDataNotFoundCode, logFields)
				return
			}
			logAndReturnError(c, http.StatusInternalServerError, err, ErrInvalidEmailDataCode, logFields)
			return
		}

		// 正常取得はINFOレベル（重要な業務イベント）
		logger.Logger.Info("EmailDataを取得しました",
			append(logFields,
				zap.String("incident_status", emailData.Incident.Status.Name))...)

		c.JSON(http.StatusOK, gin.H{
			"EmailData": emailData,
			"Incident":  emailData.Incident,
		})
	}
}
