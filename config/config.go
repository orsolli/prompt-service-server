package config

import (
	"os"
)

type Config struct {
	Port                   string
	JWTSecret              string
	CSRFTokenExpirySeconds int
	CSRFTokenSecret        string
}

func LoadConfig() *Config {
	return &Config{
		Port:                   os.Getenv("PORT"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		CSRFTokenExpirySeconds: 300, // 5 minutes
		CSRFTokenSecret:        os.Getenv("CSRF_TOKEN_SECRET"),
	}
}
