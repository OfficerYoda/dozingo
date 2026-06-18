package auth_test

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/officeryoda/dozingo/internal/auth"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate test key: %v", err)
	}
	return key
}

func testUserID() [16]byte {
	var id [16]byte
	_, _ = rand.Read(id[:])
	return id
}

func TestTOTPCipher_RoundTrip(t *testing.T) {
	key := testKey(t)
	c, err := auth.NewTOTPCipher(key)
	if err != nil {
		t.Fatalf("NewTOTPCipher: %v", err)
	}

	userID := testUserID()
	secret := "JBSWY3DPEHPK3PXP"

	sealed, err := c.Seal(userID, secret)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	if sealed == secret {
		t.Fatal("sealed output should not equal plaintext")
	}

	opened, err := c.Open(userID, sealed)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if opened != secret {
		t.Fatalf("Open returned %q, want %q", opened, secret)
	}
}

func TestTOTPCipher_DifferentNoncePerSeal(t *testing.T) {
	key := testKey(t)
	c, err := auth.NewTOTPCipher(key)
	if err != nil {
		t.Fatalf("NewTOTPCipher: %v", err)
	}

	userID := testUserID()
	secret := "JBSWY3DPEHPK3PXP"

	s1, _ := c.Seal(userID, secret)
	s2, _ := c.Seal(userID, secret)

	if s1 == s2 {
		t.Fatal("two seals of the same plaintext should produce different ciphertexts")
	}
}

func TestTOTPCipher_AADMismatch(t *testing.T) {
	key := testKey(t)
	c, err := auth.NewTOTPCipher(key)
	if err != nil {
		t.Fatalf("NewTOTPCipher: %v", err)
	}

	userID1 := testUserID()
	userID2 := testUserID()
	secret := "JBSWY3DPEHPK3PXP"

	sealed, err := c.Seal(userID1, secret)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	_, err = c.Open(userID2, sealed)
	if err == nil {
		t.Fatal("Open with wrong userID should fail")
	}
}

func TestTOTPCipher_TamperedCiphertext(t *testing.T) {
	key := testKey(t)
	c, err := auth.NewTOTPCipher(key)
	if err != nil {
		t.Fatalf("NewTOTPCipher: %v", err)
	}

	userID := testUserID()
	secret := "JBSWY3DPEHPK3PXP"

	sealed, err := c.Seal(userID, secret)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	// Decode, flip a byte in the ciphertext portion, re-encode
	raw, _ := base64.StdEncoding.DecodeString(sealed)
	raw[len(raw)-1] ^= 0xFF
	tampered := base64.StdEncoding.EncodeToString(raw)

	_, err = c.Open(userID, tampered)
	if err == nil {
		t.Fatal("Open with tampered ciphertext should fail")
	}
}

func TestTOTPCipher_UnknownVersion(t *testing.T) {
	key := testKey(t)
	c, err := auth.NewTOTPCipher(key)
	if err != nil {
		t.Fatalf("NewTOTPCipher: %v", err)
	}

	userID := testUserID()
	secret := "JBSWY3DPEHPK3PXP"

	sealed, err := c.Seal(userID, secret)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	// Set version byte to 0xFF
	raw, _ := base64.StdEncoding.DecodeString(sealed)
	raw[0] = 0xFF
	modified := base64.StdEncoding.EncodeToString(raw)

	_, err = c.Open(userID, modified)
	if err == nil {
		t.Fatal("Open with unknown version should fail")
	}
}

func TestTOTPCipher_WrongKeyLength(t *testing.T) {
	_, err := auth.NewTOTPCipher([]byte("too-short"))
	if err == nil {
		t.Fatal("NewTOTPCipher should reject keys that are not 32 bytes")
	}

	longKey := make([]byte, 64)
	_, err = auth.NewTOTPCipher(longKey)
	if err == nil {
		t.Fatal("NewTOTPCipher should reject keys that are not 32 bytes")
	}
}

func TestTOTPCipher_WrongKey(t *testing.T) {
	key1 := testKey(t)
	key2 := testKey(t)

	c1, _ := auth.NewTOTPCipher(key1)
	c2, _ := auth.NewTOTPCipher(key2)

	userID := testUserID()
	secret := "JBSWY3DPEHPK3PXP"

	sealed, err := c1.Seal(userID, secret)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	_, err = c2.Open(userID, sealed)
	if err == nil {
		t.Fatal("Open with wrong key should fail")
	}
}
