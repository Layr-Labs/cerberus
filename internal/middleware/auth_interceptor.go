package middleware

import (
	"context"
	"errors"
	"strings"

	v1 "github.com/Layr-Labs/cerberus-api/pkg/api/v1"
	"github.com/Layr-Labs/cerberus/internal/common"
	"github.com/Layr-Labs/cerberus/internal/database/repository"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor creates a selective authentication interceptor
func AuthInterceptor(
	protectedServiceName string,
	keyMetadataRepo repository.KeyMetadataRepository,
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if the current service should be protected
		if !strings.HasPrefix(info.FullMethod, "/"+protectedServiceName) {
			// Skip auth for non-protected services
			return handler(ctx, req)
		}

		// Get metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get authorization token
		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Validate the token (implement your own validation logic)
		valid, err := validateToken(ctx, authHeader[0], req, keyMetadataRepo)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		if !valid {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// If authentication successful, proceed with the handler
		return handler(ctx, req)
	}
}

// Example token validation function - replace with your own implementation
func validateToken(
	ctx context.Context,
	token string,
	req interface{},
	keyMetadataRepo repository.KeyMetadataRepository,
) (bool, error) {
	var pubKeyG1 string
	switch r := req.(type) {
	case *v1.SignGenericRequest:
		pubKeyG1 = r.GetPublicKeyG1()
	case *v1.SignG1Request:
		pubKeyG1 = r.GetPublicKeyG1()
	default:
		return false, errors.New("invalid request type")
	}

	keyMetadata, err := keyMetadataRepo.Get(ctx, pubKeyG1)
	if err != nil {
		return false, err
	}

	requestAPIKeyHash := common.CreateSHA256Hash(token)

	if keyMetadata.ApiKeyHash != requestAPIKeyHash {
		return false, errors.New("invalid token")
	}

	return true, nil
}
