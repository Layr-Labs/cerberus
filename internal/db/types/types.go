package types

import "time"

type Mapping struct {
	ApiKeyHash   string    `json:"api_key_hash"`
	PublicKeyHex string    `json:"public_key_hex"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Locked       bool      `json:"locked"`
}
