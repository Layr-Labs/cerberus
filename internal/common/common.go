package common

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

func Trim0x(s string) string {
	if len(s) >= 2 && s[0:2] == "0x" {
		return s[2:]
	}
	return s
}

func RemovePrefix(s string, prefix string) string {
	if len(s) > len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func CreateSHA256Hash(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}

func GenerateNewAPIKeyAndHash() (string, string, error) {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return "", "", err
	}
	apiKey := newUUID.String()
	apiKeyHash := CreateSHA256Hash(apiKey)
	return apiKey, apiKeyHash, nil
}
