package admin

import (
	"context"
	"log/slog"
	"time"

	v1 "github.com/Layr-Labs/cerberus-api/pkg/api/v1"
	"github.com/Layr-Labs/cerberus/internal/common"
	"github.com/Layr-Labs/cerberus/internal/configuration"
	"github.com/Layr-Labs/cerberus/internal/database/repository"
	"github.com/Layr-Labs/cerberus/internal/metrics"
)

var _ v1.AdminServer = (*Service)(nil)

type Service struct {
	config          *configuration.Configuration
	logger          *slog.Logger
	metrics         metrics.Recorder
	keyMetadataRepo repository.KeyMetadataRepository

	v1.UnimplementedAdminServer
}

func NewService(
	config *configuration.Configuration,
	logger *slog.Logger,
	metrics metrics.Recorder,
	keyMetadataRepo repository.KeyMetadataRepository,
) *Service {
	return &Service{
		config:          config,
		logger:          logger.With("component", "admin"),
		metrics:         metrics,
		keyMetadataRepo: keyMetadataRepo,
	}
}

func (s *Service) GenerateNewApiKey(
	ctx context.Context,
	req *v1.GenerateNewApiKeyRequest,
) (*v1.GenerateNewApiKeyResponse, error) {
	metadata, err := s.keyMetadataRepo.Get(ctx, req.PublicKeyG1)
	if err != nil {
		return nil, err
	}

	apiKey, apiKeyHash, err := common.GenerateNewAPIKeyAndHash()
	if err != nil {
		return nil, err
	}

	err = s.keyMetadataRepo.UpdateAPIKeyHash(ctx, metadata.PublicKeyG1, apiKeyHash)
	if err != nil {
		return nil, err
	}

	return &v1.GenerateNewApiKeyResponse{
		ApiKey:      apiKey,
		PublicKeyG1: metadata.PublicKeyG1,
	}, nil
}

func (s *Service) LockKey(
	ctx context.Context,
	req *v1.LockKeyRequest,
) (*v1.LockKeyResponse, error) {
	err := s.keyMetadataRepo.UpdateLockStatus(ctx, req.PublicKeyG1, true)
	if err != nil {
		return nil, err
	}
	return &v1.LockKeyResponse{}, nil
}

func (s *Service) UnlockKey(
	ctx context.Context,
	req *v1.UnlockKeyRequest,
) (*v1.UnlockKeyResponse, error) {
	err := s.keyMetadataRepo.UpdateLockStatus(ctx, req.PublicKeyG1, false)
	if err != nil {
		return nil, err
	}
	return &v1.UnlockKeyResponse{}, nil
}

func (s *Service) ListAllKeys(
	ctx context.Context,
	req *v1.ListAllKeysRequest,
) (*v1.ListAllKeysResponse, error) {
	keys, err := s.keyMetadataRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	response := &v1.ListAllKeysResponse{
		Keys: make([]*v1.KeyMetadata, 0, len(keys)),
	}
	for _, key := range keys {
		response.Keys = append(response.Keys, &v1.KeyMetadata{
			PublicKeyG1: key.PublicKeyG1,
			PublicKeyG2: key.PublicKeyG2,
			CreatedAt:   key.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   key.UpdatedAt.Format(time.RFC3339),
			Locked:      key.Locked,
		})
	}
	return response, nil
}
