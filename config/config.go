package config

import (
	"os"
)

type Config struct {
	Port                   string
	CSRFTokenExpirySeconds int
	CSRFTokenSecret        string
	MaxRequestBodySize     int64
}

func LoadConfig() *Config {
	return &Config{
		Port:                   os.Getenv("PORT"),
		CSRFTokenExpirySeconds: 300, // 5 minutes
		CSRFTokenSecret:        os.Getenv("CSRF_TOKEN_SECRET"),
		MaxRequestBodySize:     10 * 1024 * 1024, // 10MB limit
	}
}
