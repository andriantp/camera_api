package crypto_test

import (
	"camera_api/pkg/utils/crypto"
	"encoding/base64"
	"testing"
)

func TestGenerateKeyx(t *testing.T) {
	encoded, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	key, err := crypto.DecodeKey(encoded)
	if err != nil {
		t.Fatalf("DecodeKey() error = %v", err)
	}

	if len(key) != 32 {
		t.Fatalf("key length = %d, want 32", len(key))
	}

}

func TestGenerateKeyUnique(t *testing.T) {
	key1, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	key2, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	if key1 == key2 {
		t.Fatal("GenerateKey() generated identical keys")
	}
}

func TestDecodeKey(t *testing.T) {
	original := make([]byte, 32)

	for i := range original {
		original[i] = byte(i)
	}

	encoded := base64.StdEncoding.EncodeToString(original)

	decoded, err := crypto.DecodeKey(encoded)
	if err != nil {
		t.Fatalf("DecodeKey() error = %v", err)
	}

	if string(decoded) != string(original) {
		t.Fatal("decoded key does not match original key")
	}
}

func TestDecodeKeyInvalidBase64(t *testing.T) {
	_, err := crypto.DecodeKey("invalid-base64!!!")
	if err == nil {
		t.Fatal("DecodeKey() expected error")
	}
}

func TestDecodeKeyInvalidLength(t *testing.T) {
	key := make([]byte, 16)
	encoded := base64.StdEncoding.EncodeToString(key)

	_, err := crypto.DecodeKey(encoded)
	if err == nil {
		t.Fatal("DecodeKey() expected error")
	}
}
