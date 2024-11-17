package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProfileResponse struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")

		endpoint := os.Getenv("DB_PILOT_SERVICE_URL") + "/session-result"
		_, err := PostToDBpilotAPI(token, "", endpoint)
		if err != nil {
			if token != os.Getenv("SERVICE_TOKEN") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthrized"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func PostToDBpilotAPI(authHeader string, data string, endpoint string) (ProfileResponse, error) {

	jsonData, err := json.Marshal((data))
	if err != nil {
		return ProfileResponse{}, fmt.Errorf("failed to Marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return ProfileResponse{}, fmt.Errorf("failed to Request Body: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+authHeader)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ProfileResponse{}, fmt.Errorf("failed to Make Request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll((resp.Body))
		return ProfileResponse{}, fmt.Errorf("failed to Request Body: %d %s", resp.StatusCode, string(body))
	}

	var responsesBody ProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&responsesBody); err != nil {
		return ProfileResponse{}, fmt.Errorf("failed to Request Body: %v", err)
	}

	return responsesBody, nil
}
