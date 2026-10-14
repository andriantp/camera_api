package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateKey() (string, error) {
	key := make([]byte, 32)

	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("generate encryption key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(key), nil
}

func DecodeKey(encoded string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode encryption key: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("invalid encryption key length")
	}

	return key, nil
}
