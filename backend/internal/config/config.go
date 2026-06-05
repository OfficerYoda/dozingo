package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string `env:"DATABASE_URL,required"`
	Port         int    `env:"PORT" envDefault:"4242"`
	SecureCookie bool   `env:"SECURE_COOKIE" envDefault:"true"`

	ResendAPIKey      string `env:"RESEND_API_KEY" required:"true"`
	MailSenderAddress string `env:"MAIL_SENDER_ADDRESS" required:"true"`

	GarageEndpoint   string `env:"GARAGE_ENDPOINT" required:"true"`
	GarageAccessKey  string `env:"GARAGE_ACCESS_KEY" required:"true"`
	GarageSecretKey  string `env:"GARAGE_SECRET_KEY" required:"true"`
	GarageBucketName string `env:"GARAGE_BUCKET_NAME" envDefault:"profile-pictures"`
}

func Load() (*Config, error) {
	// Load .env file
	_ = godotenv.Load()

	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("parsing env config: %w", err)
	}

	return cfg, nil
}
