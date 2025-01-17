package kms

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"

	v1 "github.com/Layr-Labs/cerberus-api/pkg/api/v1"

	"github.com/Layr-Labs/cerberus/internal/common"
	"github.com/Layr-Labs/cerberus/internal/configuration"
	"github.com/Layr-Labs/cerberus/internal/database/model"
	"github.com/Layr-Labs/cerberus/internal/database/repository"
	"github.com/Layr-Labs/cerberus/internal/metrics"
	"github.com/Layr-Labs/cerberus/internal/store"

	"github.com/Layr-Labs/bn254-keystore-go/curve"
	"github.com/Layr-Labs/bn254-keystore-go/keystore"
	"github.com/Layr-Labs/bn254-keystore-go/mnemonic"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	config          *configuration.Configuration
	logger          *slog.Logger
	store           store.Store
	metrics         metrics.Recorder
	keyMetadataRepo repository.KeyMetadataRepository

	v1.UnimplementedKeyManagerServer
}

func NewService(
	config *configuration.Configuration,
	store store.Store,
	keyMetadataRepo repository.KeyMetadataRepository,
	logger *slog.Logger,
	metrics metrics.Recorder,
) *Service {
	return &Service{
		config:          config,
		store:           store,
		metrics:         metrics,
		keyMetadataRepo: keyMetadataRepo,
		logger:          logger.With("component", "kms"),
	}
}

func (k *Service) GenerateKeyPair(
	ctx context.Context,
	req *v1.GenerateKeyPairRequest,
) (*v1.GenerateKeyPairResponse, error) {
	password := req.GetPassword()

	// Generate a new BLS key pair
	keyPair, err := keystore.NewKeyPair(password, mnemonic.English)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to generate BLS key pair: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	g2PubKey, err := keyPair.GetG2PublicKey(curve.BN254)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to get G2 public key: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	pubKeyHex, err := k.store.StoreKey(ctx, keyPair)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to save BLS key pair to file: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Generate a new API key and hash
	apiKey, apiKeyHash, err := common.GenerateNewAPIKeyAndHash()
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to generate API key: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = k.keyMetadataRepo.Create(ctx, &model.KeyMetadata{
		PublicKeyG1: pubKeyHex,
		PublicKeyG2: g2PubKey,
		ApiKeyHash:  apiKeyHash,
	})
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to save key metadata: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert the private key to a hex string
	pkBytesSlice := make([]byte, len(keyPair.PrivateKey))
	copy(pkBytesSlice, keyPair.PrivateKey[:])
	privKeyHex := common.Trim0x(hex.EncodeToString(pkBytesSlice))

	return &v1.GenerateKeyPairResponse{
		PublicKeyG1: pubKeyHex,
		PublicKeyG2: g2PubKey,
		PrivateKey:  privKeyHex,
		Mnemonic:    keyPair.Mnemonic,
		ApiKey:      apiKey,
	}, nil
}

func (k *Service) ImportKey(
	ctx context.Context,
	req *v1.ImportKeyRequest,
) (*v1.ImportKeyResponse, error) {
	pkString := req.GetPrivateKey()
	password := req.GetPassword()
	pkMnemonic := req.GetMnemonic()
	var pkBytes []byte
	var err error

	if pkMnemonic != "" {
		ks, err := keystore.NewKeyPairFromMnemonic(pkMnemonic, password)
		if err != nil {
			k.logger.Error(fmt.Sprintf("Failed to import key pair from mnemonic: %v", err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		pkBytes = ks.PrivateKey
	} else {
		pkInt, ok := new(big.Int).SetString(pkString, 10)
		if ok {
			// It's a bigInt
			pkBytes = pkInt.Bytes()
		} else {
			// It's a hex string
			pkHex := common.Trim0x(pkString)
			pkBytes, err = hex.DecodeString(pkHex)
			if err != nil {
				k.logger.Error(fmt.Sprintf("Failed to import key pair from string: %v", err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	ks := &keystore.KeyPair{
		PrivateKey: pkBytes,
	}

	g1PubKey, err := ks.GetG1PublicKey(curve.BN254)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to get G1 public key: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	g2PubKey, err := ks.GetG2PublicKey(curve.BN254)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to get G2 public key: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = k.keyMetadataRepo.Get(ctx, g1PubKey)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "key already exists")
	}
	if err != repository.ErrKeyNotFound {
		k.logger.Error(fmt.Sprintf("Failed to get key metadata: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	pubKeyHex, err := k.store.StoreKey(
		ctx,
		&keystore.KeyPair{PrivateKey: pkBytes, Password: password},
	)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to save BLS key pair to file: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Generate a new API key and hash
	apiKey, apiKeyHash, err := common.GenerateNewAPIKeyAndHash()
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to generate API key: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = k.keyMetadataRepo.Create(ctx, &model.KeyMetadata{
		PublicKeyG1: pubKeyHex,
		PublicKeyG2: g2PubKey,
		ApiKeyHash:  apiKeyHash,
	})
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to save key metadata: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.ImportKeyResponse{
		PublicKeyG1: pubKeyHex,
		PublicKeyG2: g2PubKey,
		ApiKey:      apiKey,
	}, nil
}

func (k *Service) ListKeys(
	ctx context.Context,
	req *v1.ListKeysRequest,
) (*v1.ListKeysResponse, error) {
	keys, err := k.keyMetadataRepo.List(ctx)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to list keys: %v", err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	pubKeys := make([]*v1.PublicKey, len(keys))
	for i, key := range keys {
		pubKeys[i] = &v1.PublicKey{
			PublicKeyG1: key.PublicKeyG1,
			PublicKeyG2: key.PublicKeyG2,
		}
	}

	return &v1.ListKeysResponse{PublicKeys: pubKeys}, nil
}

func (k *Service) GetKeyMetadata(
	ctx context.Context,
	req *v1.GetKeyMetadataRequest,
) (*v1.GetKeyMetadataResponse, error) {
	metadata, err := k.keyMetadataRepo.Get(ctx, req.GetPublicKeyG1())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.GetKeyMetadataResponse{
		PublicKeyG1: metadata.PublicKeyG1,
		PublicKeyG2: metadata.PublicKeyG2,
		CreatedAt:   metadata.CreatedAt.Unix(),
		UpdatedAt:   metadata.UpdatedAt.Unix(),
	}, nil
}
