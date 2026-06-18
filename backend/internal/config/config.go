// Package config loads and exposes the typed application configuration from environment variables.
package config

import (
	"encoding/base64"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string `env:"DATABASE_URL,required"`
	Port         int    `env:"PORT" envDefault:"4242"`
	SecureCookie bool   `env:"SECURE_COOKIE" envDefault:"true"`

	// TOTPEncryptionKey is a base64-encoded 32-byte key used to encrypt TOTP
	// secrets at rest (AES-256-GCM). Generate with: openssl rand -base64 32
	TOTPEncryptionKey string `env:"TOTP_ENCRYPTION_KEY,required"`

	ResendAPIKey      string `env:"RESEND_API_KEY" required:"true"`
	MailSenderAddress string `env:"MAIL_SENDER_ADDRESS" required:"true"`

	GarageEndpoint   string `env:"GARAGE_ENDPOINT" required:"true"`
	GaragePublicURL  string `env:"GARAGE_PUBLIC_URL" required:"true"`
	GarageAccessKey  string `env:"GARAGE_ACCESS_KEY" required:"true"`
	GarageSecretKey  string `env:"GARAGE_SECRET_KEY" required:"true"`
	GarageBucketName string `env:"GARAGE_BUCKET_NAME" envDefault:"profile-pictures"`
}

// DecodeTOTPKey decodes and validates the TOTP encryption key.
// Returns the raw 32-byte key or an error if the value is invalid.
func (c *Config) DecodeTOTPKey() ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(c.TOTPEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("TOTP_ENCRYPTION_KEY: invalid base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("TOTP_ENCRYPTION_KEY: must decode to 32 bytes, got %d", len(key))
	}
	return key, nil
}

func Load() (*Config, error) {
	// Load .env file
	_ = godotenv.Load()

	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("parsing env config: %w", err)
	}

	// Validate the TOTP key early so misconfiguration fails at startup.
	if _, err := cfg.DecodeTOTPKey(); err != nil {
		return nil, err
	}

	return cfg, nil
}
