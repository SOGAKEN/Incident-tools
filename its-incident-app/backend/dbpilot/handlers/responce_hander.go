package handlers

import (
	"dbpilot/logger"
	"dbpilot/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CreateResponseRequest struct {
	IncidentID uint      `json:"incident_id"`
	Datetime   time.Time `json:"datetime"`
	Responder  string    `json:"responder"`
	Content    string    `json:"content"`
	Status     string    `json:"status"`
	Vender     int       `json:"vender"`
}

func CreateResponse(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateResponseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Logger.Warn("不正なレスポンス作成リクエスト",
				zap.Error(err),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// リクエスト情報のログ
		logger.Logger.Info("レスポンス作成リクエストを受信",
			zap.Uint("incident_id", req.IncidentID),
			zap.String("responder", req.Responder),
			zap.String("status", req.Status),
			zap.Int("vender", req.Vender),
		)

		jst, _ := time.LoadLocation("Asia/Tokyo")
		currentTime := time.Now().In(jst)

		// トランザクションを開始
		tx := db.Begin()
		if tx.Error != nil {
			logger.Logger.Error("トランザクション開始に失敗",
				zap.Error(tx.Error),
				zap.Uint("incident_id", req.IncidentID),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}

		// トランザクション処理の終了時に実行する処理を定義
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				logger.Logger.Error("パニックが発生したためロールバック",
					zap.Any("recover", r),
					zap.Uint("incident_id", req.IncidentID),
				)
			}
		}()

		// レスポンスを作成
		response := models.Response{
			IncidentID: req.IncidentID,
			Datetime:   currentTime, // Datetimeフィールドを設定
			Responder:  req.Responder,
			Content:    req.Content,
		}

		// レスポンスを保存
		if err := tx.Create(&response).Error; err != nil {
			tx.Rollback()
			logger.Logger.Error("レスポンスの作成に失敗",
				zap.Error(err),
				zap.Uint("incident_id", req.IncidentID),
				zap.String("responder", req.Responder),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response"})
			return
		}

		logger.Logger.Info("レスポンスを作成しました",
			zap.Uint("response_id", response.ID),
			zap.Uint("incident_id", req.IncidentID),
		)

		updateData := models.Incident{
			Assignee: req.Responder,
			Status:   req.Status,
			Vender:   req.Vender,
		}

		// インシデントの更新
		if err := tx.Model(&models.Incident{}).
			Where("id = ?", req.IncidentID).
			Updates(updateData).Error; err != nil {
			tx.Rollback()
			logger.Logger.Error("インシデントの更新に失敗",
				zap.Error(err),
				zap.Uint("incident_id", req.IncidentID),
				zap.String("status", req.Status),
				zap.String("assignee", req.Responder),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update incident",
			})
			return
		}

		logger.Logger.Info("インシデントを更新しました",
			zap.Uint("incident_id", req.IncidentID),
			zap.String("status", req.Status),
			zap.String("assignee", req.Responder),
		)

		// トランザクションをコミット
		if err := tx.Commit().Error; err != nil {
			logger.Logger.Error("トランザクションのコミットに失敗",
				zap.Error(err),
				zap.Uint("incident_id", req.IncidentID),
				zap.Uint("response_id", response.ID),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		logger.Logger.Info("トランザクションを正常にコミットしました",
			zap.Uint("incident_id", req.IncidentID),
			zap.Uint("response_id", response.ID),
		)

		c.JSON(http.StatusOK, gin.H{
			"message": "Response created and incident updated successfully",
			"id":      response.ID,
		})
	}
}
