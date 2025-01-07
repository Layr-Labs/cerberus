package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Layr-Labs/cerberus/internal/database/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

func TestWithContainer_KeyMetadataRepository_Create(t *testing.T) {
	testDB := SetupTestDB(t)

	metadata := &model.KeyMetadata{
		PublicKeyG1: "test_key_1",
		PublicKeyG2: "test_key_2",
	}

	err := testDB.Repo.Create(context.Background(), metadata)
	require.NoError(t, err)

	// Verify the record was created
	var result model.KeyMetadata
	err = testDB.db.QueryRow(
		"SELECT public_key_g1, public_key_g2, created_at, updated_at FROM public.keys_metadata WHERE public_key_g1 = $1",
		metadata.PublicKeyG1,
	).Scan(&result.PublicKeyG1, &result.PublicKeyG2, &result.CreatedAt, &result.UpdatedAt)

	assert.NoError(t, err)
	assert.Equal(t, metadata.PublicKeyG1, result.PublicKeyG1)
	assert.Equal(t, metadata.PublicKeyG2, result.PublicKeyG2)
	assert.WithinDuration(t, time.Now().UTC(), result.CreatedAt, 2*time.Second)
}

func TestKeyMetadataRepository_Create(t *testing.T) {
	testDB := SetupTestDB(t)
	// No need to defer db.Close() as it's handled by t.Cleanup

	tests := []struct {
		name    string
		input   *model.KeyMetadata
		wantErr bool
	}{
		{
			name: "successful creation",
			input: &model.KeyMetadata{
				PublicKeyG1: "test_key_1",
				PublicKeyG2: "test_key_2",
			},
			wantErr: false,
		},
		{
			name: "duplicate key",
			input: &model.KeyMetadata{
				PublicKeyG1: "test_key_1",
				PublicKeyG2: "test_key_2",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testDB.Repo.Create(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify the record was created
			var result model.KeyMetadata
			err = testDB.db.QueryRow(
				"SELECT public_key_g1, public_key_g2, created_at, updated_at FROM public.keys_metadata WHERE public_key_g1 = $1",
				tt.input.PublicKeyG1,
			).Scan(&result.PublicKeyG1, &result.PublicKeyG2, &result.CreatedAt, &result.UpdatedAt)

			assert.NoError(t, err)
			assert.Equal(t, tt.input.PublicKeyG1, result.PublicKeyG1)
			assert.Equal(t, tt.input.PublicKeyG2, result.PublicKeyG2)
			assert.WithinDuration(t, time.Now(), result.CreatedAt, 2*time.Second)
			assert.WithinDuration(t, time.Now(), result.UpdatedAt, 2*time.Second)
		})
	}
}

func TestKeyMetadataRepository_Get(t *testing.T) {
	testDB := SetupTestDB(t)
	// No need to defer db.Close() as it's handled by t.Cleanup

	// Create test data
	testKey := &model.KeyMetadata{
		PublicKeyG1: "test_key_1",
		PublicKeyG2: "test_key_2",
	}
	err := testDB.Repo.Create(context.Background(), testKey)
	require.NoError(t, err)

	tests := []struct {
		name      string
		keyG1     string
		wantKey   string
		wantError bool
	}{
		{
			name:      "existing key",
			keyG1:     "test_key_1",
			wantKey:   "test_key_2",
			wantError: false,
		},
		{
			name:      "non-existing key",
			keyG1:     "non_existing_key",
			wantKey:   "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testDB.Repo.Get(context.Background(), tt.keyG1)
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.keyG1, result.PublicKeyG1)
			assert.Equal(t, tt.wantKey, result.PublicKeyG2)
		})
	}
}

func TestKeyMetadataRepository_Update(t *testing.T) {
	testDB := SetupTestDB(t)
	// No need to defer db.Close() as it's handled by t.Cleanup

	// Create initial test data
	initialKey := &model.KeyMetadata{
		PublicKeyG1: "test_key_1",
		PublicKeyG2: "test_key_2",
	}
	err := testDB.Repo.Create(context.Background(), initialKey)
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   *model.KeyMetadata
		wantErr bool
	}{
		{
			name: "successful update",
			input: &model.KeyMetadata{
				PublicKeyG1: "test_key_1",
			},
			wantErr: false,
		},
		{
			name: "non-existing key",
			input: &model.KeyMetadata{
				PublicKeyG1: "non_existing_key",
				PublicKeyG2: "updated_key_2",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testDB.Repo.Update(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify the update
			result, err := testDB.Repo.Get(context.Background(), tt.input.PublicKeyG1)
			assert.NoError(t, err)
			assert.WithinDuration(t, time.Now(), result.UpdatedAt, 2*time.Second)
		})
	}
}

func TestKeyMetadataRepository_Delete(t *testing.T) {
	testDB := SetupTestDB(t)
	// No need to defer db.Close() as it's handled by t.Cleanup

	// Create test data
	testKey := &model.KeyMetadata{
		PublicKeyG1: "test_key_1",
		PublicKeyG2: "test_key_2",
	}
	err := testDB.Repo.Create(context.Background(), testKey)
	require.NoError(t, err)

	tests := []struct {
		name    string
		keyG1   string
		wantErr bool
	}{
		{
			name:    "existing key",
			keyG1:   "test_key_1",
			wantErr: false,
		},
		{
			name:    "non-existing key",
			keyG1:   "non_existing_key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testDB.Repo.Delete(context.Background(), tt.keyG1)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify the deletion
			_, err = testDB.Repo.Get(context.Background(), tt.keyG1)
			assert.Error(t, err)
		})
	}
}

func TestKeyMetadataRepository_List(t *testing.T) {
	testDB := SetupTestDB(t)
	// No need to defer db.Close() as it's handled by t.Cleanup

	// Create test data
	testKeys := []*model.KeyMetadata{
		{
			PublicKeyG1: "test_key_1",
			PublicKeyG2: "test_key_2",
		},
		{
			PublicKeyG1: "test_key_3",
			PublicKeyG2: "test_key_4",
		},
	}

	for _, key := range testKeys {
		err := testDB.Repo.Create(context.Background(), key)
		require.NoError(t, err)
	}

	t.Run("list all keys", func(t *testing.T) {
		results, err := testDB.Repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, results, len(testKeys))

		// Verify the order (should be ordered by created_at DESC)
		assert.Equal(t, "test_key_3", results[0].PublicKeyG1)
		assert.Equal(t, "test_key_1", results[1].PublicKeyG1)
	})
}
