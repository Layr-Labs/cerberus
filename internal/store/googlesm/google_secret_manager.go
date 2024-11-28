package googlesm

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"google.golang.org/api/iterator"

	"github.com/Layr-Labs/bn254-keystore-go/curve"
	"github.com/Layr-Labs/bn254-keystore-go/keystore"

	"github.com/Layr-Labs/cerberus/internal/crypto"
	"github.com/Layr-Labs/cerberus/internal/store"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

var _ store.Store = (*Keystore)(nil)

const (
	storagePrefix = "cerberus"
	ProjectKey    = "project"
)

type Keystore struct {
	smClient  *secretmanager.Client
	projectID string

	logger *slog.Logger
}

func NewKeystore(projectID string, logger *slog.Logger) (*Keystore, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Keystore{
		smClient:  client,
		projectID: projectID,
		logger:    logger.With("component", "google-secret-manager-store"),
	}, nil
}

func (k Keystore) RetrieveKey(
	ctx context.Context,
	pubKey string,
	password string,
) (*crypto.KeyPair, error) {
	// Build the request
	secretID := storagePrefix + pubKey
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", k.projectID, secretID),
	}

	// Access the secret version
	result, err := k.smClient.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to access secret version: %v", err)
	}

	privKeyHex := hex.EncodeToString(result.Payload.Data)
	kp, err := crypto.NewKeyPairFromHexString(privKeyHex)
	if err != nil {
		return nil, err
	}

	return kp, nil
}

func (k Keystore) StoreKey(ctx context.Context, keyPair *keystore.KeyPair) (string, error) {
	pubKey, err := keystore.BlsSkToPk(keyPair.PrivateKey, string(curve.BN254))
	if err != nil {
		return "", err
	}

	storageKey := storagePrefix + pubKey

	createSecretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", k.projectID),
		SecretId: storageKey,
		Secret: &secretmanagerpb.Secret{
			Labels: map[string]string{
				ProjectKey: storagePrefix,
			},
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	secret, err := k.smClient.CreateSecret(ctx, createSecretReq)
	if err != nil {
		return "", fmt.Errorf("failed to create secret: %v", err)
	}

	// Add a secret version
	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: keyPair.PrivateKey,
		},
	}

	version, err := k.smClient.AddSecretVersion(ctx, addSecretVersionReq)
	if err != nil {
		return "", fmt.Errorf("failed to add secret version: %v", err)
	}
	k.logger.Info("Stored key in secret manager with version", "version", version.Name)
	return pubKey, nil
}

func (k Keystore) ListKeys(ctx context.Context) ([]string, error) {
	filter := fmt.Sprintf("labels.%s=%s", ProjectKey, storagePrefix)
	listRequest := &secretmanagerpb.ListSecretsRequest{
		Parent: fmt.Sprintf("projects/%s", k.projectID),
		Filter: filter,
	}
	// List all secrets
	it := k.smClient.ListSecrets(ctx, listRequest)
	keys := make([]string, 0)
	for {
		secret, err := it.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("failed to list secrets: %v", err)
		}
		keys = append(keys, getPubKey(secret.GetName()))
	}
	k.logger.Debug(fmt.Sprintf("Found %d key files", len(keys)))
	return keys, nil
}

// getPubKey extracts the public key from the secret manager resource name
// The resource name is in the format:
//
//	projects/<project-id>/secrets/cerberus<pubkey>
func getPubKey(resource string) string {
	regex := fmt.Sprintf("%s(.*)", storagePrefix)

	re := regexp.MustCompile(regex)
	matches := re.FindStringSubmatch(resource)
	if len(matches) > 1 {
		result := matches[1] // Gets the content of first capturing group
		return result
	}
	return ""
}
