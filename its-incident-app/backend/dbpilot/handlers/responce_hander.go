package handlers

import (
	"dbpilot/logger"
	"dbpilot/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreateResponseRequest はレスポンス作成時のリクエスト構造体
type CreateResponseRequest struct {
	IncidentID uint      `json:"incident_id" binding:"required"`
	Datetime   time.Time `json:"datetime"`
	Responder  string    `json:"responder" binding:"required"`
	Content    string    `json:"content" binding:"required"`
	Status     string    `json:"status" binding:"required"`
	Vender     int       `json:"vender"`
}

// CreateResponse はインシデントへの返信とステータス更新を処理するハンドラー
func CreateResponse(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// リクエストのバインドと検証
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

		// 日時の設定
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

		// 現在のインシデント情報を取得
		var currentIncident models.Incident
		if err := tx.Preload("Status").First(&currentIncident, req.IncidentID).Error; err != nil {
			tx.Rollback()
			logger.Logger.Error("インシデントの取得に失敗",
				zap.Error(err),
				zap.Uint("incident_id", req.IncidentID),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}

		// 新しいステータスの取得
		var newStatus models.IncidentStatus
		if err := tx.Where("name = ?", req.Status).First(&newStatus).Error; err != nil {
			tx.Rollback()
			logger.Logger.Error("ステータスの取得に失敗",
				zap.Error(err),
				zap.String("status", req.Status),
				zap.Uint("incident_id", req.IncidentID),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("指定されたステータスが見つかりません: %v", err),
			})
			return
		}

		// ステータス遷移の記録（監査目的）
		var transition models.IncidentStatusTransition
		if err := tx.Where(
			"from_status_id = ? AND to_status_id = ? AND allowed = true",
			currentIncident.StatusID,
			newStatus.ID,
		).First(&transition).Error; err != nil {
			// 遷移が定義されていない場合は警告をログに記録
			if err == gorm.ErrRecordNotFound {
				logger.Logger.Warn("未定義のステータス遷移を許可します",
					zap.String("from_status", currentIncident.Status.Name),
					zap.String("to_status", newStatus.Name),
					zap.Uint("incident_id", req.IncidentID),
				)
			}
		}

		// レスポンスを作成
		response := models.Response{
			IncidentID: req.IncidentID,
			Datetime:   currentTime,
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

		// インシデントの更新データを準備
		updateData := models.Incident{
			Assignee: req.Responder,
			StatusID: newStatus.ID,
			Status:   newStatus,
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
			zap.String("from_status", currentIncident.Status.Name),
			zap.String("to_status", newStatus.Name),
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

		// 成功レスポンスを返す
		c.JSON(http.StatusOK, gin.H{
			"message": "Response created and incident updated successfully",
			"id":      response.ID,
			"status_change": gin.H{
				"from": currentIncident.Status.Name,
				"to":   newStatus.Name,
			},
			"incident_id": req.IncidentID,
		})
	}
}
