package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// PasswordCost is exposed so it can be reduced in tests to improve runtime
var PasswordCost = bcrypt.DefaultCost

var (
	ErrPasswordTooLong    = errors.New("password exceeds maximum length")
	ErrInvalidCredentials = errors.New("invalid credentials")
	dummyPasswordHash     string
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
		return "", ErrPasswordTooLong
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
			return ErrInvalidCredentials
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

func CheckPasswordAgainstDummy(password string) {
	_ = CheckPassword(password, dummyPasswordHash)
}
