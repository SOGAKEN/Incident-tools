package handlers

import (
	"dbpilot/models"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateResponseRequest struct {
	IncidentID uint `json:"incident_id"`
	Datetime   time.Time
	Responder  string `json:"responder"`
	Content    string `json:"content"`
	Status     string `json:"status"`
	Vender     int    `json:"vender"`
}

func CreateResponse(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateResponseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// トランザクションを開始
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}

		// トランザクション処理の終了時に実行する処理を定義
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// レスポンスを作成
		response := models.Response{
			IncidentID: req.IncidentID,
			Responder:  req.Responder,
			Content:    req.Content,
		}

		// レスポンスを保存
		if err := tx.Create(&response).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response"})
			return
		}
		updateData := models.Incident{
			Assignee: req.Responder,
			Status:   req.Status,
			Vender:   req.Vender,
		}

		// インシデントのAssigneeを更新
		if err := tx.Model(&models.Incident{}).
			Where("id = ?", req.IncidentID).
			Updates(updateData).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update incident",
			})
			return
		}

		// トランザクションをコミット
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Response created and incident updated successfully",
			"id":      response.ID,
		})
	}
}