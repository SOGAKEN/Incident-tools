package handlers

import (
	"dbpilot/models"

	"errors"

	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetIncident(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無効なIDです"})
			return
		}

		var incident models.Incident

		if err := db.Preload("Responses").
			Preload("Relations").
			Preload("Relations.RelatedIncident").
			Preload("APIData").
			First(&incident, id).Error; err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "インシデントが見つかりません"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "インシデントの取得に失敗しました"})
			}
			return
		}

		c.JSON(http.StatusOK, incident)
	}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    string `json:"code,omitempty"`
}

// GetIncidentAll は、ページネーションと関連データを含めて全てのインシデントを取得するハンドラー
func GetIncidentAll(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			incidents    []models.Incident
			total        int64
			err          error
			statusCounts []struct {
				Status string `json:"status"`
				Count  int64  `json:"count"`
			}
			req struct {
				Page   int      `json:"page"`
				Limit  int      `json:"limit"`
				Status []string `json:"status"`
				From   string   `json:"from"`
				To     string   `json:"to"`
			}
		)

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid request format",
				Code:  "INVALID_REQUEST",
			})
			return
		}

		if req.Page < 1 {
			req.Page = 1
		}
		if req.Limit <= 0 {
			req.Limit = 10
		}

		offset := (req.Page - 1) * req.Limit
		layout := "2006-01-02 15:04"
		var fromTime, toTime time.Time
		var fromProvided, toProvided bool

		if strings.TrimSpace(req.From) != "" {
			fromTime, err = time.Parse(layout, strings.TrimSpace(req.From))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "Invalid 'from' date format",
					Details: err.Error(),
					Code:    "INVALID_DATE_FORMAT",
				})
				return
			}
			fromProvided = true
		} else {
			fromTime = time.Time{}
		}

		if strings.TrimSpace(req.To) != "" {
			toTime, err = time.Parse(layout, strings.TrimSpace(req.To))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "Invalid 'to' date format",
					Details: err.Error(),
					Code:    "INVALID_DATE_FORMAT",
				})
				return
			}
			toProvided = true
		} else {
			toTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
		}

		if fromTime.After(toTime) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid date range: 'from' date is after 'to' date",
				Code:  "INVALID_DATE_RANGE",
			})
			return
		}

		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to begin transaction",
				Code:  "DB_TRANSACTION_ERROR",
			})
			return
		}

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error: "Internal server error",
					Code:  "PANIC_RECOVERED",
				})
			}
		}()

		// 有効なインシデントIDを取得するサブクエリ
		validIncidentIDs := tx.Model(&models.APIResponseData{}).
			Select("incident_id").
			Where("subject IS NOT NULL AND subject != ''")

		// メインクエリの構築
		query := tx.Model(&models.Incident{}).
			Where("id IN (?)", validIncidentIDs)

		if len(req.Status) > 0 {
			query = query.Where("status IN (?)", req.Status)
		}
		if fromProvided || toProvided {
			query = query.Where("datetime BETWEEN ? AND ?", fromTime, toTime)
		}

		// 総数カウント
		if err = query.Count(&total).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to count incidents",
				Details: err.Error(),
				Code:    "COUNT_ERROR",
			})
			return
		}

		// ステータスカウント
		statusQuery := tx.Model(&models.Incident{}).
			Where("id IN (?)", validIncidentIDs)

		// if len(req.Status) > 0 {
		// 	statusQuery = statusQuery.Where("status IN (?)", req.Status)
		// }
		// if fromProvided || toProvided {
		// 	statusQuery = statusQuery.Where("datetime BETWEEN ? AND ?", fromTime, toTime)
		// }

		if err = statusQuery.
			Select("status, count(*) as count").
			Group("status").
			Scan(&statusCounts).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to count status occurrences",
				Details: err.Error(),
				Code:    "STATUS_COUNT_ERROR",
			})
			return
		}

		// データ取得
		if err = query.Preload("Responses").
			Preload("Relations").
			Preload("Relations.RelatedIncident").
			Preload("APIData").
			Order("id DESC").
			Limit(req.Limit).
			Offset(offset).
			Find(&incidents).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to fetch incidents",
				Details: err.Error(),
				Code:    "FETCH_ERROR",
			})
			return
		}

		if err = tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "Failed to commit transaction",
				Details: err.Error(),
				Code:    "COMMIT_ERROR",
			})
			return
		}

		c.Header("Cache-Control", "private, max-age=300")

		c.JSON(http.StatusOK, gin.H{
			"data": incidents,
			"meta": gin.H{
				"total": total,
				"page":  req.Page,
				"limit": req.Limit,
				"pages": (total + int64(req.Limit) - 1) / int64(req.Limit),
			},
			"status_counts": statusCounts,
		})
	}
}