package utils

import (
	"github.com/google/uuid"
)

// GenerateSessionID セッションIDの生成
func GenerateSessionID() string {
	return uuid.New().String()
}
