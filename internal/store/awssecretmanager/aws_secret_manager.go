package awssecretmanager

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"github.com/Layr-Labs/bn254-keystore-go/curve"
	"github.com/Layr-Labs/bn254-keystore-go/keystore"

	"github.com/Layr-Labs/cerberus/internal/crypto"
	"github.com/Layr-Labs/cerberus/internal/store"
)

var _ store.Store = (*Keystore)(nil)

const storagePrefix = "cerberus/"

type Keystore struct {
	smClient *secretsmanager.Client

	logger *slog.Logger
}

func NewStoreWithEnv(
	region string,
	profile string,
	logger *slog.Logger,
) (*Keystore, error) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, err
	}

	svc := secretsmanager.NewFromConfig(cfg)
	return &Keystore{
		smClient: svc,
		logger:   logger.With("component", "aws-secret-manager-store"),
	}, nil
}

func NewStoreWithSpecifiedCredentials(
	region string,
	awsAccessKeyId string,
	awsSecretAccessKey string,
	logger *slog.Logger,
) (*Keystore, error) {
	staticCredentials := credentials.NewStaticCredentialsProvider(
		awsAccessKeyId,
		awsSecretAccessKey,
		"",
	)
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(staticCredentials),
	)

	if err != nil {
		return nil, err
	}
	svc := secretsmanager.NewFromConfig(cfg)
	return &Keystore{
		smClient: svc,
		logger:   logger.With("component", "aws-secret-manager-store"),
	}, nil
}

func (k *Keystore) RetrieveKey(
	ctx context.Context,
	pubKey string,
	password string,
) (*crypto.KeyPair, error) {
	storageKey := storagePrefix + pubKey

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     &storageKey,
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := k.smClient.GetSecretValue(ctx, input)
	if err != nil {
		return nil, err
	}

	var secretString = *result.SecretString

	kp, err := crypto.NewKeyPairFromString(secretString)
	if err != nil {
		return nil, err
	}

	return kp, nil
}

func (k *Keystore) StoreKey(ctx context.Context, keyPair *keystore.KeyPair) (string, error) {
	pubKey, err := keystore.BlsSkToPk(keyPair.PrivateKey, string(curve.BN254))
	if err != nil {
		return "", err
	}

	storageKey := storagePrefix + pubKey

	skHex := hex.EncodeToString(keyPair.PrivateKey)

	storeRequest := &secretsmanager.CreateSecretInput{
		Name:         &storageKey,
		SecretString: aws.String(skHex),
	}

	_, err = k.smClient.CreateSecret(ctx, storeRequest)
	if err != nil {
		return "", err
	}

	return pubKey, nil
}

func (k *Keystore) ListKeys(ctx context.Context) ([]string, error) {
	input := &secretsmanager.ListSecretsInput{
		Filters: []types.Filter{
			{
				Key: types.FilterNameStringTypeName,
				Values: []string{
					storagePrefix,
				},
			},
		},
	}

	result, err := k.smClient.ListSecrets(ctx, input)
	if err != nil {
		return nil, err
	}

	k.logger.Debug(fmt.Sprintf("Found %d key files", len(result.SecretList)))
	var keys []string
	for _, secret := range result.SecretList {
		keys = append(keys, removePrefix(*secret.Name, storagePrefix))
	}

	return keys, nil
}

func removePrefix(s string, prefix string) string {
	if len(s) > len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}
