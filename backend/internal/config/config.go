package config

import (
	"errors"
	"os"
	"strings"
)

// Config is the env-backed runtime configuration for the API server & CLIs.
type Config struct {
	Env  string
	Port string

	DBURL string

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool

	JWTSecret string
	HRPin     string
}

func Load() (*Config, error) {
	c := &Config{
		Env:            getenv("ENV", "development"),
		Port:           getenv("PORT", "8080"),
		DBURL:          os.Getenv("DB_URL"),
		MinioEndpoint:  getenv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:    getenv("MINIO_BUCKET", "dcgg"),
		MinioUseSSL:    strings.EqualFold(os.Getenv("MINIO_USE_SSL"), "true"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		HRPin:          os.Getenv("HR_PIN"),
	}

	var missing []string
	if c.DBURL == "" {
		missing = append(missing, "DB_URL")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.MinioAccessKey == "" {
		missing = append(missing, "MINIO_ACCESS_KEY")
	}
	if c.MinioSecretKey == "" {
		missing = append(missing, "MINIO_SECRET_KEY")
	}
	if len(missing) > 0 {
		return nil, errors.New("missing required env vars: " + strings.Join(missing, ", "))
	}
	return c, nil
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
