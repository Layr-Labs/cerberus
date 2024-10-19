package filesystem

import (
	"context"
	"encoding/hex"
	"os"
	"testing"

	"github.com/Layr-Labs/bn254-keystore-go/keystore"
	"github.com/Layr-Labs/bn254-keystore-go/mnemonic"

	"github.com/Layr-Labs/cerberus/internal/common/testutils"

	"github.com/stretchr/testify/assert"
)

const (
	tmpDir      = "tmp"
	keystoreDir = "keystore"
)

func TestFileStoreKeyOperators(t *testing.T) {
	defer cleanup()

	ctx := context.Background()
	logger := testutils.GetTestLogger()
	fs := NewStore(tmpDir+"/"+keystoreDir, logger)
	testPassword := "p@$$w0rd"

	// Run this 10 times for enough randomness to be allowed
	// in generating key pairs
	for i := 0; i < 10; i++ {
		keyPair, err := keystore.NewKeyPair(testPassword, mnemonic.English)
		assert.NoError(t, err, "Failed to generate key pair")

		pubKeyHex, err := fs.StoreKey(ctx, keyPair)
		assert.NoError(t, err, "Failed to store key")

		storedKeyPair, err := fs.RetrieveKey(ctx, pubKeyHex, testPassword)
		assert.NoError(t, err, "Failed to retrieve key")

		pubKeyBytes := storedKeyPair.PubKey.Bytes()
		pubKeySlice := pubKeyBytes[:]
		assert.Equal(t, pubKeyHex, hex.EncodeToString(pubKeySlice), "public key mismatch")

		pkBytes := storedKeyPair.PrivKey.Bytes()
		slice := pkBytes[:]
		assert.Equal(
			t,
			hex.EncodeToString(keyPair.PrivateKey),
			hex.EncodeToString(slice),
			"private key mismatch",
		)

		keys, err := fs.ListKeys(ctx)
		if err != nil {
			return
		}

		assert.Contains(t, keys, pubKeyHex, "Expected key not found")
	}
}

func cleanup() {
	err := os.RemoveAll(tmpDir)
	if err != nil {
		return
	}
}
