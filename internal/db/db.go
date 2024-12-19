package db

import "github.com/Layr-Labs/cerberus/internal/db/types"

type DB interface {
	// Get returns the value for the given key
	Get(apiKeyHash string) (*types.Mapping, error)

	// Set sets the value for the given key
	Set(apiKeyHash string, value *types.Mapping) error
}
