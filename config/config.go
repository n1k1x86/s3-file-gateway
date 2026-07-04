package config

import (
	"os"

	"github.com/joho/godotenv"
)

const (
	DEFAULT_HTTP_ADDR = "0.0.0.0:8000"
)

type Config struct {
	HTTPAddr string `env:"HTTP_ADDR"`
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	http_addr := os.Getenv("HTTP_ADDR")
	if http_addr == "" {
		http_addr = DEFAULT_HTTP_ADDR
	}

	return &Config{
		HTTPAddr: http_addr,
	}, nil
}
