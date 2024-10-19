package store

import (
	"context"

	"github.com/Layr-Labs/bn254-keystore-go/keystore"

	"github.com/Layr-Labs/cerberus/internal/crypto"
)

type Store interface {
	// RetrieveKey retrieves the private key from the store
	// using the public key and password
	// Returns the private key or an error if it fails
	// Public key is used to identify the key in the store
	RetrieveKey(ctx context.Context, pubKey string, password string) (*crypto.KeyPair, error)

	// StoreKey stores the private key in the store
	// using the public key as identifier
	// Returns an error if it fails
	// Password is used to encrypt the private key before storing if it is provided
	StoreKey(ctx context.Context, keyPair *keystore.KeyPair) (string, error)

	// ListKeys returns a list of public keys stored in the store
	ListKeys(ctx context.Context) ([]string, error)
}
