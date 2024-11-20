package awssecretmanager

import (
	"context"
	"encoding/hex"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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
	Region string

	logger *slog.Logger
}

func NewStore(region string, logger *slog.Logger) *Keystore {
	return &Keystore{
		Region: region,
		logger: logger.With("component", "aws-secret-manager-store"),
	}
}

func (k *Keystore) RetrieveKey(
	ctx context.Context,
	pubKey string,
	password string,
) (*crypto.KeyPair, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(k.Region))
	if err != nil {
		return nil, err
	}

	svc := secretsmanager.NewFromConfig(cfg)

	storageKey := storagePrefix + pubKey

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     &storageKey,
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(ctx, input)
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
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(k.Region))
	if err != nil {
		return "", err
	}

	pubKey, err := keystore.BlsSkToPk(keyPair.PrivateKey, string(curve.BN254))
	if err != nil {
		return "", err
	}

	storageKey := storagePrefix + pubKey

	svc := secretsmanager.NewFromConfig(cfg)
	skHex := hex.EncodeToString(keyPair.PrivateKey)

	storeRequest := &secretsmanager.PutSecretValueInput{
		SecretId:     &storageKey,
		SecretString: aws.String(skHex),
	}

	_, err = svc.PutSecretValue(ctx, storeRequest)
	if err != nil {
		return "", err
	}

	return pubKey, nil
}

func (k *Keystore) ListKeys(ctx context.Context) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(k.Region))
	if err != nil {
		return nil, err
	}

	svc := secretsmanager.NewFromConfig(cfg)

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

	result, err := svc.ListSecrets(ctx, input)
	if err != nil {
		return nil, err
	}

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