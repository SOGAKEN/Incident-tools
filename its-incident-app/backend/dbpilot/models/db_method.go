package models

import (
	"dbpilot/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DeleteSessionByEmail はメールアドレスに基づいてセッションを削除
func DeleteSessionByEmail(db *gorm.DB, email string) error {
	result := db.Where("email = ?", email).Delete(&LoginSession{})
	if result.Error != nil {
		logger.Logger.Error("セッション削除に失敗しました",
			zap.Error(result.Error),
			zap.String("email", email),
		)
		return result.Error
	}

	logger.Logger.Info("セッションを削除しました",
		zap.String("email", email),
		zap.Int64("deleted_count", result.RowsAffected),
	)
	return nil
}

// CreateSession は新しいセッションを作成
func CreateSession(db *gorm.DB, session *LoginSession) error {
	if err := db.Create(session).Error; err != nil {
		logger.Logger.Error("セッション作成に失敗しました",
			zap.Error(err),
			zap.String("email", session.Email),
			zap.String("session_id", session.SessionID),
		)
		return err
	}

	logger.Logger.Info("セッションを作成しました",
		zap.String("email", session.Email),
		zap.String("session_id", session.SessionID),
		zap.Time("expires_at", session.ExpiresAt),
	)
	return nil
}

// GetSessionByEmail はメールアドレスに基づいてセッションを取得
func GetSessionByEmail(db *gorm.DB, email string) (*LoginSession, error) {
	var session LoginSession
	if err := db.Where("email = ?", email).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Logger.Warn("セッションが見つかりません",
				zap.String("email", email),
			)
		} else {
			logger.Logger.Error("セッション取得に失敗しました",
				zap.Error(err),
				zap.String("email", email),
			)
		}
		return nil, err
	}

	logger.Logger.Info("セッションを取得しました",
		zap.String("email", email),
		zap.String("session_id", session.SessionID),
	)
	return &session, nil
}

// GetUserByEmail はメールアドレスに基づいてユーザーを取得
func GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Logger.Warn("ユーザーが見つかりません",
				zap.String("email", email),
			)
		} else {
			logger.Logger.Error("ユーザー取得に失敗しました",
				zap.Error(err),
				zap.String("email", email),
			)
		}
		return nil, err
	}

	logger.Logger.Info("ユーザーを取得しました",
		zap.String("email", email),
		zap.Uint("user_id", user.ID),
	)
	return &user, nil
}

// CreateUser は新しいユーザーを作成
func CreateUser(db *gorm.DB, user *User) error {
	if err := db.Create(user).Error; err != nil {
		logger.Logger.Error("ユーザー作成に失敗しました",
			zap.Error(err),
			zap.String("email", user.Email),
		)
		return err
	}

	logger.Logger.Info("ユーザーを作成しました",
		zap.String("email", user.Email),
		zap.Uint("user_id", user.ID),
	)
	return nil
}

// UpdateUser は既存のユーザー情報を更新
func UpdateUser(db *gorm.DB, user *User) error {
	if err := db.Save(user).Error; err != nil {
		logger.Logger.Error("ユーザー更新に失敗しました",
			zap.Error(err),
			zap.String("email", user.Email),
			zap.Uint("user_id", user.ID),
		)
		return err
	}

	logger.Logger.Info("ユーザー情報を更新しました",
		zap.String("email", user.Email),
		zap.Uint("user_id", user.ID),
	)
	return nil
}
