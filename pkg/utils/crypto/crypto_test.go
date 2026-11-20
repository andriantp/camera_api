package crypto_test

import (
	"camera_api/pkg/utils/crypto"
	"fmt"
	"sync"
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

func newTestCipher(t testing.TB) (*crypto.Cipher, []byte) {
	t.Helper()

	encoded, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	key, err := crypto.DecodeKey(encoded)
	if err != nil {
		t.Fatalf("DecodeKey: %v", err)
	}

	c, err := crypto.NewCipher(key)
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}

	return c, key
}

// Credentials already stored in the database were encrypted with the
// package-level Encrypt; the reusable Cipher must read them and vice versa.
func TestCipherCompatibleWithPackageFunctions(t *testing.T) {
	c, key := newTestCipher(t)

	stored, err := crypto.Encrypt("192.168.1.64", key)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if got, err := c.Decrypt(stored); err != nil || got != "192.168.1.64" {
		t.Fatalf("Cipher.Decrypt = %q, %v", got, err)
	}

	fresh, err := c.Encrypt("admin")
	if err != nil {
		t.Fatalf("Cipher.Encrypt: %v", err)
	}

	if got, err := crypto.Decrypt(fresh, key); err != nil || got != "admin" {
		t.Fatalf("Decrypt = %q, %v", got, err)
	}
}

func TestCipherConcurrentUse(t *testing.T) {
	c, _ := newTestCipher(t)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			plain := fmt.Sprintf("secret-%d", i)

			enc, err := c.Encrypt(plain)
			if err != nil {
				t.Errorf("Encrypt: %v", err)
				return
			}

			if got, err := c.Decrypt(enc); err != nil || got != plain {
				t.Errorf("Decrypt = %q, %v; want %q", got, err, plain)
			}
		}(i)
	}
	wg.Wait()
}

// Decrypting the three credential fields of one camera request.
func BenchmarkDecryptCredential(b *testing.B) {
	c, key := newTestCipher(b)
	enc, _ := c.Encrypt("192.168.1.64")

	b.Run("new cipher per call", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 3; j++ {
				_, _ = crypto.Decrypt(enc, key)
			}
		}
	})

	b.Run("reused cipher", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 3; j++ {
				_, _ = c.Decrypt(enc)
			}
		}
	})
}
