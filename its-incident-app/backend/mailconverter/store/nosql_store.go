package store

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"go.uber.org/zap"
	"mailconvertor/logger"
	"mailconvertor/models"
)

const (
	kindEmailProcessing = "EmailProcessing"
	kindServiceState    = "ServiceState"
)

type EmailStore struct {
	client    *datastore.Client
	projectID string
	logger    *zap.Logger
}

func NewEmailStore(ctx context.Context, projectID string) (*EmailStore, error) {
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &EmailStore{
		client:    client,
		projectID: projectID,
		logger:    logger.Logger,
	}, nil
}

func (s *EmailStore) CreateProcessing(ctx context.Context, messageID string) error {
	processing := &models.EmailProcessing{
		MessageID: messageID,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	key := datastore.NameKey(kindEmailProcessing, messageID, nil)
	_, err := s.client.Put(ctx, key, processing)
	if err != nil {
		return err
	}

	// 初期サービス状態の作成
	state := &models.ServiceState{
		MessageID:   messageID,
		ServiceType: "mail-converter",
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	stateKey := datastore.NameKey(kindServiceState, messageID+":mail-converter", nil)
	_, err = s.client.Put(ctx, stateKey, state)
	return err
}

func (s *EmailStore) UpdateProcessing(ctx context.Context, processing *models.EmailProcessing) error {
	processing.UpdatedAt = time.Now()
	key := datastore.NameKey(kindEmailProcessing, processing.MessageID, nil)
	_, err := s.client.Put(ctx, key, processing)
	return err
}

func (s *EmailStore) UpdateServiceState(ctx context.Context, state *models.ServiceState) error {
	state.UpdatedAt = time.Now()
	state.TruncateEmailBody() // 保存前に本文を制限

	key := datastore.NameKey(kindServiceState, state.MessageID+":"+state.ServiceType, nil)
	_, err := s.client.Put(ctx, key, state)
	return err
}

func (s *EmailStore) GetProcessing(ctx context.Context, messageID string) (*models.EmailProcessing, error) {
	key := datastore.NameKey(kindEmailProcessing, messageID, nil)
	processing := new(models.EmailProcessing)
	if err := s.client.Get(ctx, key, processing); err != nil {
		return nil, err
	}
	processing.MessageID = messageID
	return processing, nil
}

func (s *EmailStore) GetServiceState(ctx context.Context, messageID string) (*models.ServiceState, error) {
	key := datastore.NameKey(kindServiceState, messageID+":mail-converter", nil)
	state := new(models.ServiceState)
	if err := s.client.Get(ctx, key, state); err != nil {
		return nil, err
	}
	state.MessageID = messageID
	return state, nil
}

func (s *EmailStore) SetError(ctx context.Context, messageID, errorCode, errorMessage string) error {
	// 全体の状態を更新
	processing, err := s.GetProcessing(ctx, messageID)
	if err != nil {
		return err
	}
	processing.Status = models.StatusError
	if err := s.UpdateProcessing(ctx, processing); err != nil {
		return err
	}

	// サービスの状態を更新
	state, err := s.GetServiceState(ctx, messageID)
	if err != nil {
		return err
	}
	state.Status = models.StatusError
	state.ErrorCode = errorCode
	state.ErrorMessage = errorMessage
	return s.UpdateServiceState(ctx, state)
}

func (s *EmailStore) Close() {
	if s.client != nil {
		s.client.Close()
	}
}
