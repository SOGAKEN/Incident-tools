package store

import (
	"context"
	"fmt"
	"time"

	"autopilot/logger"
	"autopilot/models"
	"cloud.google.com/go/datastore"
	"go.uber.org/zap"
)

// EmailStore はメール処理の状態管理を担当する構造体です。
// Datastoreを使用して、処理の全体状態と詳細な状態を管理します。
type EmailStore struct {
	client    *datastore.Client
	projectID string
	logger    *zap.Logger
}

const (
	// Datastoreのエンティティ種別を定義します
	kindEmailProcessing = "EmailProcessing" // メール処理全体の状態用
	kindServiceState    = "ServiceState"    // 各サービスの状態用
)

// NewEmailStore は新しいEmailStoreインスタンスを作成します。
// Datastoreクライアントの初期化と基本設定を行います。
func NewEmailStore(ctx context.Context, projectID string) (*EmailStore, error) {
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create datastore client: %v", err)
	}

	store := &EmailStore{
		client:    client,
		projectID: projectID,
		logger:    logger.Logger,
	}

	store.logger.Info("EmailStoreを初期化しました",
		zap.String("project_id", projectID))

	return store, nil
}

// CreateProcessing は新しいメール処理エントリを作成します。
// EmailProcessingとServiceStateの両方のエントリを初期化します。
func (s *EmailStore) CreateProcessing(ctx context.Context, messageID string) error {
	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("operation", "CreateProcessing"),
	}

	// EmailProcessingエントリの作成
	processing := &models.EmailProcessing{
		MessageID: messageID,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	key := datastore.NameKey(kindEmailProcessing, messageID, nil)
	if _, err := s.client.Put(ctx, key, processing); err != nil {
		s.logger.Error("EmailProcessingの作成に失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to create email processing: %v", err)
	}

	// ServiceStateエントリの作成
	state := &models.ServiceState{
		MessageID:   messageID,
		ServiceType: models.ServiceAutoPilot,
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	stateKey := s.createServiceStateKey(messageID)
	if _, err := s.client.Put(ctx, stateKey, state); err != nil {
		s.logger.Error("ServiceStateの作成に失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to create service state: %v", err)
	}

	s.logger.Info("処理状態を作成しました", logFields...)
	return nil
}

// GetProcessing は指定されたメール処理の状態を取得します。
func (s *EmailStore) GetProcessing(ctx context.Context, messageID string) (*models.EmailProcessing, error) {
	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("operation", "GetProcessing"),
	}

	key := datastore.NameKey(kindEmailProcessing, messageID, nil)
	processing := new(models.EmailProcessing)

	if err := s.client.Get(ctx, key, processing); err != nil {
		if err == datastore.ErrNoSuchEntity {
			s.logger.Debug("EmailProcessingが見つかりません", logFields...)
			return nil, nil
		}
		s.logger.Error("EmailProcessingの取得に失敗しました",
			append(logFields, zap.Error(err))...)
		return nil, fmt.Errorf("failed to get email processing: %v", err)
	}

	processing.MessageID = messageID
	return processing, nil
}

// GetServiceState は指定されたサービスの状態を取得します。
func (s *EmailStore) GetServiceState(ctx context.Context, messageID string) (*models.ServiceState, error) {
	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("operation", "GetServiceState"),
	}

	key := s.createServiceStateKey(messageID)
	state := new(models.ServiceState)

	if err := s.client.Get(ctx, key, state); err != nil {
		if err == datastore.ErrNoSuchEntity {
			s.logger.Debug("ServiceStateが見つかりません", logFields...)
			return nil, nil
		}
		s.logger.Error("ServiceStateの取得に失敗しました",
			append(logFields, zap.Error(err))...)
		return nil, fmt.Errorf("failed to get service state: %v", err)
	}

	state.MessageID = messageID
	return state, nil
}

// UpdateProcessing はメール処理の状態を更新します。
func (s *EmailStore) UpdateProcessing(ctx context.Context, processing *models.EmailProcessing) error {
	logFields := []zap.Field{
		zap.String("message_id", processing.MessageID),
		zap.String("status", string(processing.Status)),
		zap.String("operation", "UpdateProcessing"),
	}

	processing.UpdatedAt = time.Now()
	key := datastore.NameKey(kindEmailProcessing, processing.MessageID, nil)

	if _, err := s.client.Put(ctx, key, processing); err != nil {
		s.logger.Error("EmailProcessingの更新に失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to update email processing: %v", err)
	}

	s.logger.Debug("EmailProcessingを更新しました", logFields...)
	return nil
}

// UpdateServiceState はサービスの状態を更新します。
func (s *EmailStore) UpdateServiceState(ctx context.Context, state *models.ServiceState) error {
	logFields := []zap.Field{
		zap.String("message_id", state.MessageID),
		zap.String("status", string(state.Status)),
		zap.String("operation", "UpdateServiceState"),
	}

	state.UpdatedAt = time.Now()
	key := s.createServiceStateKey(state.MessageID)

	if _, err := s.client.Put(ctx, key, state); err != nil {
		s.logger.Error("ServiceStateの更新に失敗しました",
			append(logFields, zap.Error(err))...)
		return fmt.Errorf("failed to update service state: %v", err)
	}

	s.logger.Debug("ServiceStateを更新しました", logFields...)
	return nil
}

// SetError はエラー状態を設定します。
// EmailProcessingとServiceStateの両方のエントリを更新します。
func (s *EmailStore) SetError(ctx context.Context, messageID string, errorCode, errorMessage string) error {
	logFields := []zap.Field{
		zap.String("message_id", messageID),
		zap.String("error_code", errorCode),
		zap.String("operation", "SetError"),
	}

	// EmailProcessingの更新
	processing, err := s.GetProcessing(ctx, messageID)
	if err != nil {
		return err
	}
	if processing != nil {
		processing.SetError(errorMessage)
		if err := s.UpdateProcessing(ctx, processing); err != nil {
			return err
		}
	}

	// ServiceStateの更新
	state, err := s.GetServiceState(ctx, messageID)
	if err != nil {
		return err
	}
	if state != nil {
		state.SetError(errorCode, errorMessage)
		if err := s.UpdateServiceState(ctx, state); err != nil {
			return err
		}
	}

	s.logger.Info("エラー状態を設定しました", logFields...)
	return nil
}

// createServiceStateKey はServiceStateのDatastoreキーを生成します。
func (s *EmailStore) createServiceStateKey(messageID string) *datastore.Key {
	return datastore.NameKey(kindServiceState, fmt.Sprintf("%s:auto-pilot", messageID), nil)
}

// Close はEmailStoreのリソースを解放します。
func (s *EmailStore) Close() {
	if s.client != nil {
		s.client.Close()
	}
}
