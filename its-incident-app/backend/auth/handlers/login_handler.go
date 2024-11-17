// auth-service/handlers/login_handler.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"auth/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type QueryUserResponse struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginUser(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// DB Pilot Serviceからユーザー情報を取得
	baseURL := os.Getenv("DB_PILOT_SERVICE_URL")
	userData := map[string]string{"email": req.Email}
	userDataJSON, _ := json.Marshal(userData)
	resp, err := http.Post(baseURL+"/login", "application/json", bytes.NewBuffer(userDataJSON))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var userResponse QueryUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
		return
	}

	// パスワード検証
	if err := bcrypt.CompareHashAndPassword([]byte(userResponse.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// セッションIDの生成
	sessionID := utils.GenerateSessionID()
	expirationTime := time.Now().Add(24 * time.Hour) // セッションの有効期限

	// セッション情報をDB Pilot Serviceに保存
	saveSessionReq := map[string]interface{}{
		"user_id":    userResponse.ID,
		"email":      userResponse.Email,
		"session_id": sessionID,
		"expires_at": expirationTime,
	}
	saveSessionReqJSON, _ := json.Marshal(saveSessionReq)
	_, err = http.Post(baseURL+"/sessions", "application/json", bytes.NewBuffer(saveSessionReqJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	// セッションIDをHTTPOnlyクッキーとしてクライアントに返す
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Path:     "/",
		Expires:  expirationTime,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}
