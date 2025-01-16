package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Layr-Labs/cerberus/internal/database/model"
	"github.com/Layr-Labs/cerberus/internal/database/repository"
)

type keyMetadataRepo struct {
	db *sql.DB
}

func NewKeyMetadataRepository(db *sql.DB) repository.KeyMetadataRepository {
	return &keyMetadataRepo{
		db: db,
	}
}

const (
	createKeyMetadataQuery = `
        INSERT INTO public.keys_metadata (
            public_key_g1, public_key_g2, created_at, updated_at, api_key_hash
        ) VALUES ($1, $2, $3, $4, $5)
    `

	getKeyMetadataQuery = `
        SELECT public_key_g1, public_key_g2, created_at, updated_at, api_key_hash, locked
        FROM public.keys_metadata
        WHERE public_key_g1 = $1
    `

	updateKeyMetadataQuery = `
        UPDATE public.keys_metadata
        SET updated_at = $1
        WHERE public_key_g1 = $2
    `

	updateAPIKeyHashQuery = `
        UPDATE public.keys_metadata
        SET api_key_hash = $1, updated_at = $2
        WHERE public_key_g1 = $3
    `

	deleteKeyMetadataQuery = `
        DELETE FROM public.keys_metadata
        WHERE public_key_g1 = $1
    `

	listKeyMetadataQuery = `
        SELECT public_key_g1, public_key_g2, created_at, updated_at
        FROM public.keys_metadata
        ORDER BY created_at DESC
    `
)

func (r *keyMetadataRepo) Create(ctx context.Context, metadata *model.KeyMetadata) error {
	if metadata.PublicKeyG1 == "" {
		return errors.New("public key g1 is required")
	}
	if metadata.PublicKeyG2 == "" {
		return errors.New("public key g2 is required")
	}

	now := time.Now().UTC()
	metadata.CreatedAt = now
	metadata.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, createKeyMetadataQuery,
		metadata.PublicKeyG1,
		metadata.PublicKeyG2,
		metadata.CreatedAt,
		metadata.UpdatedAt,
		metadata.ApiKeyHash,
	)
	return err
}

func (r *keyMetadataRepo) Get(ctx context.Context, publicKeyG1 string) (*model.KeyMetadata, error) {
	metadata := &model.KeyMetadata{}
	err := r.db.QueryRowContext(ctx, getKeyMetadataQuery, publicKeyG1).Scan(
		&metadata.PublicKeyG1,
		&metadata.PublicKeyG2,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
		&metadata.ApiKeyHash,
		&metadata.Locked,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("key metadata not found")
	}
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func (r *keyMetadataRepo) Update(ctx context.Context, metadata *model.KeyMetadata) error {
	if metadata.PublicKeyG1 == "" {
		return errors.New("public key g1 is required")
	}

	metadata.UpdatedAt = time.Now().UTC()
	result, err := r.db.ExecContext(ctx, updateKeyMetadataQuery,
		metadata.UpdatedAt,
		metadata.PublicKeyG1,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("key metadata not found")
	}
	return nil
}

func (r *keyMetadataRepo) UpdateAPIKeyHash(ctx context.Context, metadata *model.KeyMetadata) error {
	if metadata.PublicKeyG1 == "" {
		return errors.New("public key g1 is required")
	}
	if metadata.ApiKeyHash == "" {
		return errors.New("api key hash is required")
	}

	metadata.UpdatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, updateAPIKeyHashQuery,
		metadata.ApiKeyHash,
		metadata.UpdatedAt,
		metadata.PublicKeyG1,
	)
	return err
}

func (r *keyMetadataRepo) Delete(ctx context.Context, publicKeyG1 string) error {
	if publicKeyG1 == "" {
		return errors.New("public key g1 is required")
	}

	result, err := r.db.ExecContext(ctx, deleteKeyMetadataQuery, publicKeyG1)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("key metadata not found")
	}
	return nil
}

func (r *keyMetadataRepo) List(ctx context.Context) ([]*model.KeyMetadata, error) {
	rows, err := r.db.QueryContext(ctx, listKeyMetadataQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metadata []*model.KeyMetadata
	for rows.Next() {
		m := &model.KeyMetadata{}
		err := rows.Scan(
			&m.PublicKeyG1,
			&m.PublicKeyG2,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return metadata, nil
}
