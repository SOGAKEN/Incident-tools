package handlers

import (
	"dbpilot/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateIncident(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var apiRequest models.APIRequest
		if err := c.ShouldBindJSON(&apiRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
			return
		}

		// JSONデータを文字列として保存するため、再度マーシャル
		rawJSON, err := json.Marshal(apiRequest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request", "details": err.Error()})
			return
		}

		// statusがsucceededでない場合はエラーログに保存
		if apiRequest.Data.Status != "succeeded" {
			errorLog := models.ErrorLog{
				TaskID:        apiRequest.TaskID,
				WorkflowRunID: apiRequest.WorkflowRunID,
				WorkflowID:    apiRequest.Data.WorkflowID,
				Status:        apiRequest.Data.Status,
				RawJSON:       string(rawJSON),
			}

			if err := db.Create(&errorLog).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to create error log",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Error log created successfully",
				"id":      errorLog.ID,
			})
			return
		}

		// 以下、status == "succeeded" の場合の処理
		datetime := time.Unix(apiRequest.Data.CreatedAt, 0)
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// インシデントの作成
		incident := models.Incident{
			Datetime: datetime,
			Status:   "未着手",
			Assignee: "-",
			Vender:   0,
		}

		if err := tx.Create(&incident).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create incident",
				"details": err.Error(),
			})
			return
		}

		// WorkflowLogsを文字列として取得
		workflowLogsJSON, err := json.Marshal(apiRequest.Data.Outputs.WorkflowLogs)
		if err != nil {
			workflowLogsJSON = []byte("[]") // デフォルト値を設定
		}

		// API応答データの作成（構造化されたデータ）
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
			Sender:     apiRequest.Data.Outputs.Sender,
			Final:     apiRequest.Data.Outputs.Final,

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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create API response data",
				"details": err.Error(),
			})
			return
		}

		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to commit transaction",
				"details": err.Error(),
			})
			return
		}

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

type CreateIncidentRelationRequest struct {
	IncidentID        uint `json:"incident_id"`
	RelatedIncidentID uint `json:"related_incident_id"`
}

func CreateIncidentRelation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateIncidentRelationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		relation := models.IncidentRelation{
			IncidentID:        req.IncidentID,
			RelatedIncidentID: req.RelatedIncidentID,
		}

		if err := db.Create(&relation).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident relation"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Incident relation created successfully", "id": relation.ID})
	}
}
