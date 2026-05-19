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
}

func Load() (*Config, error) {
	// Load .env file
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parsing env config: %w", err)
	}

	return cfg, nil
}
