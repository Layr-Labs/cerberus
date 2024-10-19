package filesystem

import (
	"context"
	"fmt"

	"log/slog"
	"os"

	"github.com/Layr-Labs/cerberus/internal/crypto"
	"github.com/Layr-Labs/cerberus/internal/store"

	"github.com/Layr-Labs/bn254-keystore-go/curve"
	"github.com/Layr-Labs/bn254-keystore-go/keystore"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

const keyFileExtension = ".json"

var _ store.Store = (*FileStore)(nil)

type FileStore struct {
	keystoreDir string

	logger *slog.Logger
}

func NewStore(
	keystoreDir string,
	logger *slog.Logger,
) *FileStore {
	logger = logger.With("component", "filesystem-store")
	if err := os.MkdirAll(keystoreDir, 0755); err != nil {
		logger.Error(fmt.Sprintf("Error creating keystore directory: %v", err))
		os.Exit(1)
	}
	logger.Info("Created keystore directory successfully")
	return &FileStore{
		keystoreDir: keystoreDir,
		logger:      logger,
	}
}

func (s *FileStore) RetrieveKey(
	ctx context.Context,
	pubKey string,
	password string,
) (*crypto.KeyPair, error) {
	path := s.keystoreDir + "/" + pubKey + ".json"
	return readPrivateKeyFromFile(path, password)
}

func (s *FileStore) StoreKey(
	ctx context.Context,
	keyPair *keystore.KeyPair,
) (string, error) {
	keyStore, err := keyPair.Encrypt(keystore.KDFScrypt, curve.BN254)
	if err != nil {
		return "", err
	}

	err = keyStore.SaveWithPubKeyHex(s.keystoreDir, "")
	if err != nil {
		return "", err
	}

	return keyStore.PubKey, nil
}

func (s *FileStore) ListKeys(ctx context.Context) ([]string, error) {
	files, err := os.ReadDir(s.keystoreDir)
	if err != nil {
		return nil, err
	}

	s.logger.Debug(fmt.Sprintf("Found %d key files", len(files)))
	pubKeys := make([]string, len(files))
	for i, file := range files {
		pubKeys[i] = file.Name()[0 : len(file.Name())-len(keyFileExtension)]
	}

	return pubKeys, nil
}

func readPrivateKeyFromFile(path string, password string) (*crypto.KeyPair, error) {
	ks := new(keystore.Keystore)
	err := ks.FromFile(path)
	if err != nil {
		return nil, err
	}

	skBytes, err := ks.Decrypt(password)
	if err != nil {
		return nil, err
	}

	privKey := new(fr.Element).SetBytes(skBytes)
	keyPair := crypto.NewKeyPair(privKey)
	return keyPair, nil
}
