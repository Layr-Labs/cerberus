package kms

import (
	"context"
	"encoding/hex"
	"os"
	"testing"

	"github.com/Layr-Labs/bn254-keystore-go/keystore"
	"github.com/Layr-Labs/bn254-keystore-go/mnemonic"
	v1 "github.com/Layr-Labs/cerberus-api/pkg/api/v1"

	"github.com/Layr-Labs/cerberus/internal/common/testutils"
	"github.com/Layr-Labs/cerberus/internal/configuration"
	"github.com/Layr-Labs/cerberus/internal/database/repository/postgres"
	"github.com/Layr-Labs/cerberus/internal/metrics"
	"github.com/Layr-Labs/cerberus/internal/store/filesystem"

	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
)

const testPassword = "p@$$w0rd"

func setup(t *testing.T) (*Service, *filesystem.FileStore, func()) {
	logger := testutils.GetTestLogger()
	config := &configuration.Configuration{
		KeystoreDir: "testdata/keystore",
	}
	fs := filesystem.NewStore(config.KeystoreDir, logger)
	noopMetrics := metrics.NewNoopRPCMetrics()
	testDB := postgres.SetupTestDB(t)
	service := NewService(config, fs, testDB.Repo, logger, noopMetrics)
	cleanup := func() {
		os.RemoveAll("testdata")
	}

	return service, fs, cleanup
}

func TestCreateKey(t *testing.T) {
	service, fs, cleanup := setup(t)
	defer cleanup()

	ctx := context.Background()

	createResp, err := service.GenerateKeyPair(
		ctx,
		&v1.GenerateKeyPairRequest{Password: testPassword},
	)
	assert.NoError(t, err)

	storedKeyPair, err := fs.RetrieveKey(ctx, createResp.PublicKeyG1, testPassword)
	assert.NoError(t, err)

	pubKeyBytes := storedKeyPair.PubKey.Bytes()
	pubKeyHex := hex.EncodeToString(pubKeyBytes[:])
	assert.Equal(t, createResp.PublicKeyG1, pubKeyHex)

	privBytes := storedKeyPair.PrivKey.Bytes()
	privKeyHex := hex.EncodeToString(privBytes[:])
	assert.Equal(t, createResp.PrivateKey, privKeyHex)
}

func TestImportKey(t *testing.T) {
	service, fs, cleanup := setup(t)
	defer cleanup()

	ctx := context.Background()

	keyPair, err := keystore.NewKeyPair(testPassword, mnemonic.English)
	assert.NoError(t, err)

	importResp, err := service.ImportKey(ctx, &v1.ImportKeyRequest{
		PrivateKey: hex.EncodeToString(keyPair.PrivateKey),
		Password:   testPassword,
	})
	assert.NoError(t, err)

	storedKeyPair, err := fs.RetrieveKey(ctx, importResp.PublicKeyG1, testPassword)
	assert.NoError(t, err)
	privKeyBytes := storedKeyPair.PrivKey.Bytes()
	assert.Equal(t, hex.EncodeToString(keyPair.PrivateKey), hex.EncodeToString(privKeyBytes[:]))
}

func TestListKeys(t *testing.T) {
	service, fs, cleanup := setup(t)
	defer cleanup()

	ctx := context.Background()

	createResp, err := service.GenerateKeyPair(
		ctx,
		&v1.GenerateKeyPairRequest{Password: testPassword},
	)
	assert.NoError(t, err)

	listResp, err := service.ListKeys(ctx, &v1.ListKeysRequest{})
	assert.NoError(t, err)
	assert.Equal(t, listResp.PublicKeys[0].PublicKeyG1, createResp.PublicKeyG1)

	storedKeys, err := fs.ListKeys(ctx)
	assert.NoError(t, err)
	assert.Contains(t, storedKeys, createResp.PublicKeyG1)
}
