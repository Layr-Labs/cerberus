package signing

import (
	"sync"

	"github.com/Layr-Labs/cerberus/internal/crypto"
)

type KeyStoreMap struct {
	sync.Map
}

func (k *KeyStoreMap) Load(key string) (*crypto.KeyPair, bool) {
	value, ok := k.Map.Load(key)
	if !ok {
		return nil, false
	}
	return value.(*crypto.KeyPair), true
}

func (k *KeyStoreMap) Store(key string, value *crypto.KeyPair) {
	k.Map.Store(key, value)
}
