package handlers

import (
	"dbpilot/logger"
	"dbpilot/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CreateIncidentRelationRequest struct {
	IncidentID        uint `json:"incident_id"`
	RelatedIncidentID uint `json:"related_incident_id"`
}

func CreateIncident(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "CreateIncident"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		var apiRequest models.APIRequest
		if err := c.ShouldBindJSON(&apiRequest); err != nil {
			logger.Logger.Error("リクエストのバインドに失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
			return
		}

		logFields = append(logFields,
			zap.String("task_id", apiRequest.TaskID),
			zap.String("message_id", apiRequest.MessageID), // AIResponsePayloadから取得
			zap.String("workflow_run_id", apiRequest.WorkflowRunID))

		// JSONデータを文字列として保存
		rawJSON, err := json.Marshal(apiRequest)
		if err != nil {
			logger.Logger.Error("リクエストのJSONエンコードに失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request", "details": err.Error()})
			return
		}

		// statusがsucceededでない場合はエラーログに保存
		if apiRequest.Data.Status != "succeeded" {
			logger.Logger.Warn("ワークフローが失敗しました",
				append(logFields,
					zap.String("status", apiRequest.Data.Status),
					zap.String("workflow_id", apiRequest.Data.WorkflowID))...)

			errorLog := models.ErrorLog{
				TaskID:        apiRequest.TaskID,
				WorkflowRunID: apiRequest.WorkflowRunID,
				WorkflowID:    apiRequest.Data.WorkflowID,
				Status:        apiRequest.Data.Status,
				MessageID:     apiRequest.MessageID,
				RawJSON:       string(rawJSON),
			}

			if err := db.Create(&errorLog).Error; err != nil {
				logger.Logger.Error("エラーログの保存に失敗しました",
					append(logFields, zap.Error(err))...)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to create error log",
					"details": err.Error(),
				})
				return
			}

			logger.Logger.Info("エラーログを保存しました",
				append(logFields, zap.Uint("error_log_id", errorLog.ID))...)
			c.JSON(http.StatusOK, gin.H{
				"message": "Error log created successfully",
				"id":      errorLog.ID,
			})
			return
		}

		// 成功時の処理
		datetime := time.Unix(apiRequest.Data.CreatedAt, 0)
		tx := db.Begin()
		if tx.Error != nil {
			logger.Logger.Error("トランザクションの開始に失敗しました",
				append(logFields, zap.Error(tx.Error))...)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				logger.Logger.Error("パニックが発生しました",
					append(logFields, zap.Any("recover", r))...)
			}
		}()

		// インシデントの作成
		incident := models.Incident{
			Datetime:  datetime,
			Status:    "未着手",
			Assignee:  "-",
			Vender:    0,
			MessageID: apiRequest.MessageID,
		}

		if err := tx.Create(&incident).Error; err != nil {
			tx.Rollback()
			logger.Logger.Error("インシデントの作成に失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create incident",
				"details": err.Error(),
			})
			return
		}

		// WorkflowLogsの処理
		workflowLogsJSON, err := json.Marshal(apiRequest.Data.Outputs.WorkflowLogs)
		if err != nil {
			logger.Logger.Warn("ワークフローログのJSONエンコードに失敗しました",
				append(logFields, zap.Error(err))...)
			workflowLogsJSON = []byte("[]")
		}

		// API応答データの作成
		apiData := models.APIResponseData{
			IncidentID:    incident.ID,
			TaskID:        apiRequest.TaskID,
			WorkflowRunID: apiRequest.WorkflowRunID,
			WorkflowID:    apiRequest.Data.WorkflowID,
			Status:        apiRequest.Data.Status,

			Body:         apiRequest.Data.Outputs.Body,
			User:         apiRequest.Data.Outputs.User,
			WorkflowLogs: string(workflowLogsJSON),
			Host:         apiRequest.Data.Outputs.Host,
			Priority:     apiRequest.Data.Outputs.Priority,
			Subject:      apiRequest.Data.Outputs.Subject,
			From:         apiRequest.Data.Outputs.From,
			Place:        apiRequest.Data.Outputs.Place,
			IncidentText: apiRequest.Data.Outputs.Incident,
			Time:         apiRequest.Data.Outputs.Time,
			Judgment:     apiRequest.Data.Outputs.Judgment,
			Sender:       apiRequest.Data.Outputs.Sender,
			Final:        apiRequest.Data.Outputs.Final,

			ElapsedTime: apiRequest.Data.ElapsedTime,
			TotalTokens: apiRequest.Data.TotalTokens,
			TotalSteps:  apiRequest.Data.TotalSteps,
			CreatedAt:   apiRequest.Data.CreatedAt,
			FinishedAt:  apiRequest.Data.FinishedAt,
			Error:       fmt.Sprintf("%v", apiRequest.Data.Error),
			RawResponse: string(rawJSON),
		}

		if err := tx.Create(&apiData).Error; err != nil {
			tx.Rollback()
			logger.Logger.Error("API応答データの作成に失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create API response data",
				"details": err.Error(),
			})
			return
		}

		if err := tx.Commit().Error; err != nil {
			logger.Logger.Error("トランザクションのコミットに失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to commit transaction",
				"details": err.Error(),
			})
			return
		}

		logger.Logger.Info("インシデントを作成しました",
			append(logFields,
				zap.Uint("incident_id", incident.ID),
				zap.String("subject", apiData.Subject))...)

		c.JSON(http.StatusOK, gin.H{
			"message": "Incident created successfully",
			"id":      incident.ID,
			"data": gin.H{
				"incident": incident,
				"api_data": apiData,
			},
		})
	}
}

func CreateIncidentRelation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		logFields := []zap.Field{
			zap.String("handler", "CreateIncidentRelation"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}

		var req CreateIncidentRelationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Logger.Error("リクエストのバインドに失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		logFields = append(logFields,
			zap.Uint("incident_id", req.IncidentID),
			zap.Uint("related_incident_id", req.RelatedIncidentID))

		relation := models.IncidentRelation{
			IncidentID:        req.IncidentID,
			RelatedIncidentID: req.RelatedIncidentID,
		}

		if err := db.Create(&relation).Error; err != nil {
			logger.Logger.Error("インシデント関連の作成に失敗しました",
				append(logFields, zap.Error(err))...)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident relation"})
			return
		}

		logger.Logger.Info("インシデント関連を作成しました",
			append(logFields, zap.Uint("relation_id", relation.ID))...)

		c.JSON(http.StatusOK, gin.H{"message": "Incident relation created successfully", "id": relation.ID})
	}
}
