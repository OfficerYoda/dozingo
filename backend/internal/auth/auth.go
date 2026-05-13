package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// PasswordCost is exposed so it can be reduced in tests to improve runtime
var PasswordCost = bcrypt.DefaultCost

var (
	ErrPasswordTooLong    = errors.New("password exceeds maximum length")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

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
