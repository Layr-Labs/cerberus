package kms

import (
	"context"
	"encoding/hex"
	"testing"

	v1 "github.com/Layr-Labs/cerberus-api/pkg/api/v1"

	"github.com/Layr-Labs/cerberus/internal/common/testutils"
	"github.com/Layr-Labs/cerberus/internal/configuration"
	"github.com/Layr-Labs/cerberus/internal/metrics"
	"github.com/Layr-Labs/cerberus/internal/store/filesystem"

	"github.com/stretchr/testify/assert"
)

const testPassword = "p@$$w0rd"

func setup() (*Service, *filesystem.FileStore) {
	logger := testutils.GetTestLogger()
	config := &configuration.Configuration{
		KeystoreDir: "testdata/keystore",
	}
	fs := filesystem.NewStore(config.KeystoreDir, logger)
	noopMetrics := metrics.NewNoopRPCMetrics()
	service := NewService(config, fs, logger, noopMetrics)

	return service, fs
}

func TestCreateKey(t *testing.T) {
	service, fs := setup()

	ctx := context.Background()

	createResp, err := service.GenerateKeyPair(
		ctx,
		&v1.GenerateKeyPairRequest{Password: testPassword},
	)
	assert.NoError(t, err)

	storedKeyPair, err := fs.RetrieveKey(ctx, createResp.PublicKey, testPassword)
	assert.NoError(t, err)

	pubKeyBytes := storedKeyPair.PubKey.Bytes()
	pubKeyHex := hex.EncodeToString(pubKeyBytes[:])
	assert.Equal(t, createResp.PublicKey, pubKeyHex)

	privBytes := storedKeyPair.PrivKey.Bytes()
	privKeyHex := hex.EncodeToString(privBytes[:])
	assert.Equal(t, createResp.PrivateKey, privKeyHex)
}

func TestImportKey(t *testing.T) {
	service, _ := setup()

	ctx := context.Background()

	createResp, err := service.GenerateKeyPair(
		ctx,
		&v1.GenerateKeyPairRequest{Password: testPassword},
	)
	assert.NoError(t, err)

	importResp, err := service.ImportKey(ctx, &v1.ImportKeyRequest{
		PrivateKey: createResp.PrivateKey,
		Password:   testPassword,
	})
	assert.NoError(t, err)
	assert.Equal(t, createResp.PublicKey, importResp.PublicKey)
}

func TestListKeys(t *testing.T) {
	service, fs := setup()

	ctx := context.Background()

	createResp, err := service.GenerateKeyPair(
		ctx,
		&v1.GenerateKeyPairRequest{Password: testPassword},
	)
	assert.NoError(t, err)

	listResp, err := service.ListKeys(ctx, &v1.ListKeysRequest{})
	assert.NoError(t, err)
	assert.Contains(t, listResp.PublicKeys, createResp.PublicKey)

	storedKeys, err := fs.ListKeys(ctx)
	assert.NoError(t, err)
	assert.Contains(t, storedKeys, createResp.PublicKey)
}
