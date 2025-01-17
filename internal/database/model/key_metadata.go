package model

import "time"

type KeyMetadata struct {
	PublicKeyG1 string    `db:"public_key_g1"`
	PublicKeyG2 string    `db:"public_key_g2"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	ApiKeyHash  string    `db:"api_key_hash"`
	Locked      bool      `db:"locked"`
}
