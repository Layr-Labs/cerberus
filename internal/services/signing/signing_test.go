package signing

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

func TestSigning(t *testing.T) {
	// private key: 0x040ad69253b921aca71dd714cccc3095576fbe1a21f86c9b10cb5b119b1c6899
	pubKeyHex := "a3111a2232584734d526d62cbb7c9a0d4ce1984a92b7ecb85bde8878fea5d1b0"
	password := "p@$$w0rd"
	expectedSig := "0fea882fc5c936c304b0d79f4c256dbb2d38a2df74b44aaa483dfa87f1a86ede0bbc32080db378a408b90af7e264b9768a4b2f16c6953ec2611a13bc448d27e4"
	data := []byte("somedata")
	var bytes [32]byte
	copy(bytes[:], data)

	config := &configuration.Configuration{
		KeystoreDir: "testdata/keystore",
	}
	logger := testutils.GetTestLogger()
	store := filesystem.NewStore(config.KeystoreDir, logger)
	m := metrics.NewNoopRPCMetrics()
	signingService := NewService(config, store, logger, m)

	resp, err := signingService.SignGeneric(context.Background(), &v1.SignGenericRequest{
		PublicKeyG1: pubKeyHex,
		Data:        bytes[:],
		Password:    password,
	})
	assert.NoError(t, err)
	assert.Equal(t, expectedSig, hex.EncodeToString(resp.Signature))
}

func TestSigningG1(t *testing.T) {
	// private key: 0x040ad69253b921aca71dd714cccc3095576fbe1a21f86c9b10cb5b119b1c6899
	pubKeyHex := "a3111a2232584734d526d62cbb7c9a0d4ce1984a92b7ecb85bde8878fea5d1b0"
	password := "p@$$w0rd"
	expectedSig := "24a87f9eab63a40c62831d2e9598e698f8819b15093c268b89c1a521f7d986650000000000000000000000000000000000000000000000000000000000000000"
	data := []byte("somedata")
	var bytes [64]byte
	copy(bytes[:], data)

	config := &configuration.Configuration{
		KeystoreDir: "testdata/keystore",
	}
	logger := testutils.GetTestLogger()
	store := filesystem.NewStore(config.KeystoreDir, logger)
	m := metrics.NewNoopRPCMetrics()
	signingService := NewService(config, store, logger, m)

	resp, err := signingService.SignG1(context.Background(), &v1.SignG1Request{
		PublicKeyG1: pubKeyHex,
		Data:        bytes[:],
		Password:    password,
	})
	assert.NoError(t, err)
	assert.Equal(t, expectedSig, hex.EncodeToString(resp.Signature))
}
