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
	ResendAPIKey string `env:"RESEND_API_KEY" required:"true"`
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
