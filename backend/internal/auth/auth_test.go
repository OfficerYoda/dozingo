package auth

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func init() {
	// Run unit tests with minimum bcrypt cost for speed.
	PasswordCost = bcrypt.MinCost
}

func TestHashPassword_RoundTrip(t *testing.T) {
	hash, err := HashPassword("hunter2")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == "hunter2" {
		t.Fatal("HashPassword returned the plaintext")
	}
	if err := CheckPassword("hunter2", hash); err != nil {
		t.Fatalf("CheckPassword should accept the original password, got: %v", err)
	}
}

func TestHashPassword_DifferentSaltEachTime(t *testing.T) {
	a, err := HashPassword("samepw")
	if err != nil {
		t.Fatalf("first HashPassword failed: %v", err)
	}
	b, err := HashPassword("samepw")
	if err != nil {
		t.Fatalf("second HashPassword failed: %v", err)
	}
	if a == b {
		t.Fatal("expected different hashes for the same password (different salts)")
	}
}

func TestHashPassword_TooLong(t *testing.T) {
	tooLong := strings.Repeat("a", 73)
	_, err := HashPassword(tooLong)
	if !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("expected ErrPasswordTooLong, got: %v", err)
	}
}

func TestHashPassword_AtLimit(t *testing.T) {
	atLimit := strings.Repeat("a", 72)
	hash, err := HashPassword(atLimit)
	if err != nil {
		t.Fatalf("expected 72-byte password to be accepted, got: %v", err)
	}
	if err := CheckPassword(atLimit, hash); err != nil {
		t.Fatalf("CheckPassword on 72-byte password failed: %v", err)
	}
}

func TestCheckPassword_Wrong(t *testing.T) {
	hash, err := HashPassword("correct")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	err = CheckPassword("incorrect", hash)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestCheckPassword_MalformedHash(t *testing.T) {
	err := CheckPassword("anything", "not-a-bcrypt-hash")
	if err == nil {
		t.Fatal("expected an error for a malformed hash")
	}
	if errors.Is(err, ErrInvalidCredentials) {
		t.Fatal("malformed-hash error must not be reported as ErrInvalidCredentials")
	}
}

func TestGenerateSessionToken_Format(t *testing.T) {
	token := GenerateSessionToken()
	if token == "" {
		t.Fatal("GenerateSessionToken returned empty string")
	}
	// 32 bytes URL-safe base64-encoded -> 44 chars including padding.
	if len(token) != 44 {
		t.Fatalf("expected 44-char token, got %d (%q)", len(token), token)
	}
	if _, err := base64.URLEncoding.DecodeString(token); err != nil {
		t.Fatalf("token %q is not valid URL-safe base64: %v", token, err)
	}
}

func TestGenerateSessionToken_Uniqueness(t *testing.T) {
	const n = 1000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		tok := GenerateSessionToken()
		if _, dup := seen[tok]; dup {
			t.Fatalf("duplicate token after %d iterations: %q", i, tok)
		}
		seen[tok] = struct{}{}
	}
}
