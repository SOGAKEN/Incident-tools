package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"notification/models"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func NotifyHandler(c *gin.Context) {

	var req models.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	teamsWebhookURL := os.Getenv("TEAMS_WEBHOOK_URL")
	if teamsWebhookURL == "" {
		RespondWithError(c, http.StatusInternalServerError, "Teams webhook URL not configured")
		return
	}

	if err := SendTeamsNotification(teamsWebhookURL, req); err != nil {
		RespondWithError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to send notification: %v", err))
		return
	}

	authHeader := c.GetHeader("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	endpoint := os.Getenv("DB_PILOT_SERVICE_URL") + "/responses"

	_, err := SendDBpilot(req, token, endpoint)
	if err != nil {
		fmt.Printf("db pilot error: %V\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification sent successfully",
		"status":  "success",
	})
}

func SendTeamsNotification(webhookURL string, notification models.NotificationRequest) error {
	teamsReq := map[string]interface{}{
		"title":   notification.Title,
		"content": notification.Content,
	}

	teamsReqJSON, err := json.Marshal(teamsReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(teamsReqJSON))
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("teams webhook returned unexpected status: %d", resp.StatusCode)
	}

	return nil
}

func RespondWithError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func SendDBpilot(req models.NotificationRequest, authHeader string, endpoint string) (string, error) {
	dbReq, err := json.Marshal(req)
	if err != nil {
		return "failed", fmt.Errorf(("failed to marshal DB pilot request: %v"), err)
	}
	dbReqBody := bytes.NewBuffer(dbReq)

	dbClient := &http.Client{}
	dbRequest, err := http.NewRequest("POST", endpoint, dbReqBody)
	if err != nil {
		return "failed", fmt.Errorf(("failed to marshal DB pilot request: %v"), err)
	}
	dbRequest.Header.Set("Content-Type", "application/json")
	dbRequest.Header.Set("Authorization", "Bearer "+authHeader)

	dbResp, err := dbClient.Do(dbRequest)
	if err != nil {
		return "failed", fmt.Errorf(("failed to marshal DB pilot request: %v"), err)
	}
	defer dbResp.Body.Close()

	if dbResp.StatusCode != http.StatusOK {
		return "failed", fmt.Errorf(("failed to marshal DB pilot request: %d"), dbResp.StatusCode)
	}

	return "success", nil
}
