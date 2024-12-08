package handlers

import (
	"dbpilot/logger"
	"dbpilot/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CreateIncidentRelationRequest struct {
	SourceMessageID string `json:"source_message_id" binding:"required"`
	TargetMessageID string `json:"target_message_id" binding:"required"`
}

// CreateIncidentRelation はインシデント間の関係を作成するハンドラー
func CreateIncidentRelation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateIncidentRelationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Logger.Error("リクエストのバインドに失敗しました",
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request format",
			})
			return
		}

		logger.Logger.Info("インシデント関係の作成を開始します",
			zap.String("source_message_id", req.SourceMessageID),
			zap.String("target_message_id", req.TargetMessageID),
		)

		// トランザクション開始
		err := db.Transaction(func(tx *gorm.DB) error {
			// 元となるインシデントを検索
			var sourceIncident models.Incident
			if err := tx.Where("message_id = ?", req.SourceMessageID).First(&sourceIncident).Error; err != nil {
				logger.Logger.Error("ソースインシデントの取得に失敗しました",
					zap.Error(err),
					zap.String("message_id", req.SourceMessageID),
				)
				return err
			}

			// 関連付け先のインシデントを検索
			var targetIncident models.Incident
			if err := tx.Where("message_id = ?", req.TargetMessageID).First(&targetIncident).Error; err != nil {
				logger.Logger.Error("ターゲットインシデントの取得に失敗しました",
					zap.Error(err),
					zap.String("message_id", req.TargetMessageID),
				)
				return err
			}

			// 既存の関係をチェック
			var existingRelation models.IncidentRelation
			err := tx.Where("incident_id = ? AND related_incident_id = ?",
				sourceIncident.ID, targetIncident.ID).First(&existingRelation).Error

			if err == nil {
				logger.Logger.Warn("既に関係が存在します",
					zap.Uint("source_id", sourceIncident.ID),
					zap.Uint("target_id", targetIncident.ID),
				)
				return gorm.ErrDuplicatedKey
			} else if err != gorm.ErrRecordNotFound {
				return err
			}

			// 関係を作成
			relation := models.IncidentRelation{
				IncidentID:        sourceIncident.ID,
				RelatedIncidentID: targetIncident.ID,
			}

			if err := tx.Create(&relation).Error; err != nil {
				logger.Logger.Error("関係の作成に失敗しました",
					zap.Error(err),
					zap.Uint("source_id", sourceIncident.ID),
					zap.Uint("target_id", targetIncident.ID),
				)
				return err
			}

			logger.Logger.Info("インシデント関係を作成しました",
				zap.Uint("relation_id", relation.ID),
				zap.Uint("source_id", sourceIncident.ID),
				zap.Uint("target_id", targetIncident.ID),
			)

			return nil
		})

		if err != nil {
			statusCode := http.StatusInternalServerError
			message := "Internal server error"

			if err == gorm.ErrRecordNotFound {
				statusCode = http.StatusNotFound
				message = "Incident not found"
			} else if err == gorm.ErrDuplicatedKey {
				statusCode = http.StatusConflict
				message = "Relation already exists"
			}

			c.JSON(statusCode, gin.H{
				"error": message,
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Incident relation created successfully",
		})
	}
}
