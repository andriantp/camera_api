package crypto_test

import (
	"camera_api/pkg/utils/crypto"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	plainText := "camera-password"

	encrypted, err := crypto.Encrypt(plainText, key)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := crypto.Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plainText {
		t.Fatalf("Decrypt() = %q, want %q", decrypted, plainText)
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	key := make([]byte, 32)
	plainText := "camera-password"

	encrypted1, err := crypto.Encrypt(plainText, key)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	encrypted2, err := crypto.Encrypt(plainText, key)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if encrypted1 == encrypted2 {
		t.Fatal("Encrypt() produced identical ciphertext")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key := make([]byte, 32)
	wrongKey := make([]byte, 32)

	wrongKey[0] = 1

	encrypted, err := crypto.Encrypt("camera-password", key)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	_, err = crypto.Decrypt(encrypted, wrongKey)
	if err == nil {
		t.Fatal("Decrypt() expected error with wrong key")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key := make([]byte, 32)

	_, err := crypto.Decrypt("invalid-ciphertext", key)
	if err == nil {
		t.Fatal("Decrypt() expected error")
	}
}

func TestDecryptShortCiphertext(t *testing.T) {
	key := make([]byte, 32)

	_, err := crypto.Decrypt("YWJj", key)
	if err == nil {
		t.Fatal("Decrypt() expected error for short ciphertext")
	}
}

func TestEncryptEmptyPlaintext(t *testing.T) {
	key := make([]byte, 32)

	encrypted, err := crypto.Encrypt("", key)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := crypto.Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != "" {
		t.Fatalf("Decrypt() = %q, want empty string", decrypted)
	}
}
