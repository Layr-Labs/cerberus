package signing

import (
	"context"
	"fmt"
	"log/slog"

	v1 "github.com/Layr-Labs/cerberus-api/pkg/api/v1"

	"github.com/Layr-Labs/cerberus/internal/common"
	"github.com/Layr-Labs/cerberus/internal/configuration"
	"github.com/Layr-Labs/cerberus/internal/crypto"
	"github.com/Layr-Labs/cerberus/internal/metrics"
	"github.com/Layr-Labs/cerberus/internal/store"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	config   *configuration.Configuration
	logger   *slog.Logger
	store    store.Store
	metrics  metrics.Recorder
	keyCache map[string]*crypto.KeyPair
	v1.UnimplementedSignerServer
}

func NewService(
	config *configuration.Configuration,
	store store.Store,
	logger *slog.Logger,
	metrics metrics.Recorder,
) *Service {
	return &Service{
		config:   config,
		store:    store,
		metrics:  metrics,
		logger:   logger.With("component", "signing"),
		keyCache: make(map[string]*crypto.KeyPair),
	}
}

func (s *Service) SignGeneric(
	ctx context.Context,
	req *v1.SignGenericRequest,
) (*v1.SignGenericResponse, error) {
	// Take the public key and data from the request
	pubKeyHex := common.Trim0x(req.GetPublicKeyG1())
	password := req.GetPassword()

	if _, ok := s.keyCache[pubKeyHex]; !ok {
		s.logger.Info(fmt.Sprintf("In memory cache miss. Retrieving key for %s", pubKeyHex))
		blsKey, err := s.store.RetrieveKey(ctx, pubKeyHex, password)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to retrieve key: %v", err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		s.keyCache[pubKeyHex] = blsKey
	}
	blsKey := s.keyCache[pubKeyHex]

	data := req.GetData()
	if len(data) > 32 {
		s.logger.Error("Data is too long, must be 32 bytes")
		return nil, status.Error(codes.InvalidArgument, "data is too long, must be 32 bytes")
	}

	var byteArray [32]byte
	copy(byteArray[:], data)
	// Sign the data with the private key
	sig := blsKey.SignMessage(byteArray)
	s.logger.Info(fmt.Sprintf("Signed a message successfully using %s", pubKeyHex))
	return &v1.SignGenericResponse{Signature: sig.Serialize()}, nil
}

func (s *Service) SignG1(
	ctx context.Context,
	req *v1.SignG1Request,
) (*v1.SignG1Response, error) {
	// Take the public key and data from the request
	pubKeyHex := common.Trim0x(req.GetPublicKeyG1())
	password := req.GetPassword()

	if _, ok := s.keyCache[pubKeyHex]; !ok {
		s.logger.Info(fmt.Sprintf("In memory cache miss. Retrieving key for %s", pubKeyHex))
		blsKey, err := s.store.RetrieveKey(ctx, pubKeyHex, password)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to retrieve key: %v", err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		s.keyCache[pubKeyHex] = blsKey
	}
	blsKey := s.keyCache[pubKeyHex]

	g1Bytes := req.GetData()
	g1Point := new(crypto.G1Point)
	g1Point = g1Point.Deserialize(g1Bytes)

	sig := blsKey.SignHashedToCurveMessage(g1Point.G1Affine)
	return &v1.SignG1Response{Signature: sig.Serialize()}, nil
}
