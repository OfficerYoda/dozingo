package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

const (
	totpCipherVersion byte = 0x01
	nonceSize              = 12 // standard GCM nonce
	keySize                = 32 // AES-256
)

// TOTPCipher encrypts and decrypts TOTP secrets using AES-256-GCM.
//
// The stored ciphertext format (before base64 encoding) is:
//
//	[1-byte version][12-byte nonce][ciphertext + 16-byte GCM tag]
//
// Additional authenticated data (AAD) is the raw 16-byte user UUID,
// binding the ciphertext to the owning row and preventing row-swapping attacks.
type TOTPCipher struct {
	aead cipher.AEAD
}

// NewTOTPCipher creates a cipher from a 32-byte AES-256 key.
func NewTOTPCipher(key []byte) (*TOTPCipher, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("totp cipher key must be %d bytes, got %d", keySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return &TOTPCipher{aead: gcm}, nil
}

// Seal encrypts a TOTP secret and returns a base64-encoded ciphertext string
// suitable for storage in the database. userID is the raw 16-byte UUID used as AAD.
func (c *TOTPCipher) Seal(userID [16]byte, plaintext string) (string, error) {
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := c.aead.Seal(nil, nonce, []byte(plaintext), userID[:])

	// format: version || nonce || ciphertext+tag
	out := make([]byte, 0, 1+nonceSize+len(ciphertext))
	out = append(out, totpCipherVersion)
	out = append(out, nonce...)
	out = append(out, ciphertext...)

	return base64.StdEncoding.EncodeToString(out), nil
}

// Open decrypts a base64-encoded ciphertext previously produced by Seal.
// userID must be the same raw 16-byte UUID that was passed to Seal.
func (c *TOTPCipher) Open(userID [16]byte, encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	if len(raw) < 1+nonceSize+c.aead.Overhead() {
		return "", errors.New("ciphertext too short")
	}

	version := raw[0]
	if version != totpCipherVersion {
		return "", fmt.Errorf("unsupported cipher version: %d", version)
	}

	nonce := raw[1 : 1+nonceSize]
	ciphertext := raw[1+nonceSize:]

	plaintext, err := c.aead.Open(nil, nonce, ciphertext, userID[:])
	if err != nil {
		return "", fmt.Errorf("decrypt totp secret: %w", err)
	}

	return string(plaintext), nil
}
