package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	DEFAULT_HTTP_ADDR = "0.0.0.0:8000"
)

type Config struct {
	HTTPAddr   string `env:"HTTP_ADDR"`
	S3Key      string `env:"S3_KEY"`
	S3Secret   string `env:"S3_SECRET"`
	S3Region   string `env:"S3_REGION"`
	S3Endpoint string `env:"S3_ENDPOINT"`
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = DEFAULT_HTTP_ADDR
	}

	s3Key := os.Getenv("S3_KEY")
	if s3Key == "" {
		return nil, fmt.Errorf("s3 key is empty")
	}

	s3Secret := os.Getenv("S3_SECRET")
	if s3Secret == "" {
		return nil, fmt.Errorf("s3 secret is empty")
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		return nil, fmt.Errorf("s3 region is empty")
	}

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint == "" {
		return nil, fmt.Errorf("s3 endpoint is empty")
	}

	return &Config{
		HTTPAddr:   httpAddr,
		S3Key:      s3Key,
		S3Secret:   s3Secret,
		S3Region:   s3Region,
		S3Endpoint: s3Endpoint,
	}, nil
}
