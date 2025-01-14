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
	config  *configuration.Configuration
	logger  *slog.Logger
	store   store.Store
	metrics metrics.Recorder
	keyMap  KeyStoreMap
	v1.UnimplementedSignerServer
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
		logger:  logger.With("component", "signing"),
		keyMap:  KeyStoreMap{},
	}
}

func (s *Service) SignGeneric(
	ctx context.Context,
	req *v1.SignGenericRequest,
) (*v1.SignGenericResponse, error) {
	// Take the public key and data from the request
	pubKeyHex := common.Trim0x(req.GetPublicKeyG1())
	password := req.GetPassword()

	if _, ok := s.keyMap.Load(pubKeyHex); !ok {
		s.logger.Info(fmt.Sprintf("In memory cache miss. Retrieving key for %s", pubKeyHex))
		blsKey, err := s.store.RetrieveKey(ctx, pubKeyHex, password)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to retrieve key: %v", err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		s.keyMap.Store(pubKeyHex, blsKey)
	}
	blsKey, _ := s.keyMap.Load(pubKeyHex)

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
	signatureBytes := sig.RawBytes()
	return &v1.SignGenericResponse{Signature: signatureBytes[:]}, nil
}

func (s *Service) SignG1(
	ctx context.Context,
	req *v1.SignG1Request,
) (*v1.SignG1Response, error) {
	// Take the public key and data from the request
	pubKeyHex := common.Trim0x(req.GetPublicKeyG1())
	password := req.GetPassword()

	if pubKeyHex == "" {
		return nil, status.Error(codes.InvalidArgument, "public key is required")
	}

	g1Bytes := req.GetData()
	if len(g1Bytes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "data must be > 0 bytes")
	}

	if _, ok := s.keyMap.Load(pubKeyHex); !ok {
		s.logger.Info(fmt.Sprintf("In memory cache miss. Retrieving key for %s", pubKeyHex))
		blsKey, err := s.store.RetrieveKey(ctx, pubKeyHex, password)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to retrieve key: %v", err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		s.keyMap.Store(pubKeyHex, blsKey)
	}
	blsKey, _ := s.keyMap.Load(pubKeyHex)

	g1Point := new(crypto.G1Point)
	g1Point = g1Point.Deserialize(g1Bytes)

	sig := blsKey.SignHashedToCurveMessage(g1Point.G1Affine)
	s.logger.Info(fmt.Sprintf("Signed a G1 message successfully using %s", pubKeyHex))
	signatureBytes := sig.RawBytes()
	return &v1.SignG1Response{Signature: signatureBytes[:]}, nil
}
