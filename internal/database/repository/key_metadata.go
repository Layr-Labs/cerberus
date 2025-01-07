package repository

import (
	"context"

	"github.com/Layr-Labs/cerberus/internal/database/model"
)

type KeyMetadataRepository interface {
	Create(ctx context.Context, metadata *model.KeyMetadata) error
	Get(ctx context.Context, publicKeyG1 string) (*model.KeyMetadata, error)
	Update(ctx context.Context, metadata *model.KeyMetadata) error
	Delete(ctx context.Context, publicKeyG1 string) error
	List(ctx context.Context) ([]*model.KeyMetadata, error)
}
