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
	"github.com/Layr-Labs/cerberus/internal/metrics"
	"github.com/Layr-Labs/cerberus/internal/store"

	"github.com/Layr-Labs/bn254-keystore-go/keystore"
	"github.com/Layr-Labs/bn254-keystore-go/mnemonic"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	config  *configuration.Configuration
	logger  *slog.Logger
	store   store.Store
	metrics metrics.Recorder

	v1.UnimplementedKeyManagerServer
}

func NewService(
	config *configuration.Configuration,
	store store.Store,
	logger *slog.Logger,
	metrics metrics.Recorder,
) *Service {
	return &Service{
		config:  config,
		store:   store,
		metrics: metrics,
		logger:  logger.With("component", "kms"),
	}
}

func (k *Service) GenerateKeyPair(
	ctx context.Context,
	req *v1.GenerateKeyPairRequest,
) (*v1.GenerateKeyPairResponse, error) {
	observe := k.metrics.RecordRPCServerRequest("kms/GenerateKeyPair")
	defer observe()
	password := req.GetPassword()

	// Generate a new BLS key pair
	keyPair, err := keystore.NewKeyPair(password, mnemonic.English)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to generate BLS key pair: %v", err))
		k.metrics.RecordRPCServerResponse("kms/GenerateKeyPair", metrics.FailureLabel)
		return nil, status.Error(codes.Internal, err.Error())
	}

	pubKeyHex, err := k.store.StoreKey(ctx, keyPair)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to save BLS key pair to file: %v", err))
		k.metrics.RecordRPCServerResponse("kms/GenerateKeyPair", metrics.FailureLabel)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert the private key to a hex string
	pkBytesSlice := make([]byte, len(keyPair.PrivateKey))
	copy(pkBytesSlice, keyPair.PrivateKey[:])
	privKeyHex := common.Trim0x(hex.EncodeToString(pkBytesSlice))

	k.metrics.RecordRPCServerResponse("kms/GenerateKeyPair", metrics.SuccessLabel)
	return &v1.GenerateKeyPairResponse{
		PublicKey:  pubKeyHex,
		PrivateKey: privKeyHex,
		Mnemonic:   keyPair.Mnemonic,
	}, nil
}

func (k *Service) ImportKey(
	ctx context.Context,
	req *v1.ImportKeyRequest,
) (*v1.ImportKeyResponse, error) {
	observe := k.metrics.RecordRPCServerRequest("kms/ImportKey")
	defer observe()
	pkString := req.GetPrivateKey()
	password := req.GetPassword()
	pkMnemonic := req.GetMnemonic()
	var pkBytes []byte
	var err error

	if pkMnemonic != "" {
		ks, err := keystore.NewKeyPairFromMnemonic(pkMnemonic, password)
		if err != nil {
			k.logger.Error(fmt.Sprintf("Failed to import key pair from mnemonic: %v", err))
			k.metrics.RecordRPCServerResponse("kms/ImportKey", metrics.FailureLabel)
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
				k.metrics.RecordRPCServerResponse("kms/ImportKey", metrics.FailureLabel)
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	pubKeyHex, err := k.store.StoreKey(
		ctx,
		&keystore.KeyPair{PrivateKey: pkBytes, Password: password},
	)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to save BLS key pair to file: %v", err))
		k.metrics.RecordRPCServerResponse("kms/ImportKey", metrics.FailureLabel)
		return nil, status.Error(codes.Internal, err.Error())
	}

	k.metrics.RecordRPCServerResponse("kms/ImportKey", metrics.SuccessLabel)
	return &v1.ImportKeyResponse{PublicKey: pubKeyHex}, nil
}

func (k *Service) ListKeys(
	ctx context.Context,
	req *v1.ListKeysRequest,
) (*v1.ListKeysResponse, error) {
	observe := k.metrics.RecordRPCServerRequest("kms/ListKeys")
	defer observe()
	pubKeys, err := k.store.ListKeys(ctx)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Failed to list keys: %v", err))
		k.metrics.RecordRPCServerResponse("kms/ListKeys", metrics.FailureLabel)
		return nil, status.Error(codes.Internal, err.Error())
	}

	k.metrics.RecordRPCServerResponse("kms/ListKeys", metrics.SuccessLabel)
	return &v1.ListKeysResponse{PublicKeys: pubKeys}, nil
}
