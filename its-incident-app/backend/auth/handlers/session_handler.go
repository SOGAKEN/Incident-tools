// auth-service/handlers/session_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func VerifySession(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
		return
	}

	

	authHeader := c.GetHeader("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	endpoint := os.Getenv("DB_PILOT_SERVICE_URL") + "/sessions"

	_, err := SendDBpilot(token, endpoint)
	if err != nil {
		fmt.Printf("db pilot error: %V\n", err)
	}


	// 有効なトークン
	c.JSON(http.StatusOK, gin.H{"message": "Token is valid"})
}

func SendDBpilot(authHeader string, endpoint string) (string, error) {
	

	dbClient := &http.Client{}
	dbRequest, err := http.NewRequest("GET", endpoint, nil)
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
