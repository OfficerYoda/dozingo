// Package auth provides password hashing, token generation, and related primitives used by the auth flows.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/officeryoda/dozingo/internal/domain"
)

const recoveryCodeCount = 10

// PasswordCost is exposed so it can be reduced in tests to improve runtime
var (
	PasswordCost      = bcrypt.DefaultCost
	dummyPasswordHash string
)

func init() {
	// Dummy hash for timing attack prevention
	hash, err := HashPassword("dummy-password")
	if err != nil {
		panic("failed to generate dummy hash")
	}
	dummyPasswordHash = hash
}

func HashPassword(password string) (string, error) {
	// bcrypt has a 72 byte limit, anything beyond is silently truncated
	if len(password) > 72 {
		return "", fmt.Errorf("password to long: %w", domain.ErrUnprocessableEntity)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CheckPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return domain.ErrUnauthorized
		}

		return err
	}

	return nil
}

func GenerateToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "" // rand.Read never returns an error
	}

	return base64.URLEncoding.EncodeToString(bytes)
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func CheckPasswordAgainstDummy(password string) {
	_ = CheckPassword(password, dummyPasswordHash)
}

// GenerateRecoveryCodes generates codes in format: XXXXX-XXXXX
func GenerateRecoveryCodes() ([]string, error) {
	codes := make([]string, recoveryCodeCount)

	for i := range codes {
		b := make([]byte, 8)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("generating recovery code: %w", err)
		}
		codes[i] = fmt.Sprintf("%08X-%08X", b[:4], b[4:])
	}

	return codes, nil
}
