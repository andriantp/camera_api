package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Cipher encrypts and decrypts strings with AES-GCM. Create it once and
// reuse it: building the AES block and GCM mode is the costly part, and a
// cipher.AEAD is safe for concurrent use.
type Cipher struct {
	aead cipher.AEAD
}

// NewCipher returns a Cipher for a 16, 24 or 32 byte key.
func NewCipher(key []byte) (*Cipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &Cipher{aead: aead}, nil
}

// Encrypt returns base64(nonce || ciphertext) with a fresh random nonce.
func (c *Cipher) Encrypt(plainText string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := c.aead.Seal(nonce, nonce, []byte(plainText), nil)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt reverses Encrypt.
func (c *Cipher) Decrypt(cipherText string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	nonceSize := c.aead.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("invalid ciphertext")
	}

	plainText, err := c.aead.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

// Encrypt encrypts with a one-off Cipher. Prefer reusing a Cipher.
func Encrypt(plainText string, key []byte) (string, error) {
	c, err := NewCipher(key)
	if err != nil {
		return "", err
	}

	return c.Encrypt(plainText)
}

// Decrypt decrypts with a one-off Cipher. Prefer reusing a Cipher.
func Decrypt(cipherText string, key []byte) (string, error) {
	c, err := NewCipher(key)
	if err != nil {
		return "", err
	}

	return c.Decrypt(cipherText)
}
