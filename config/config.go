package config

import (
    "os"
)

type Config struct {
    Port string
    JWTSecret string
}

func LoadConfig() *Config {
    return &Config{
        Port: os.Getenv("PORT"),
        JWTSecret: os.Getenv("JWT_SECRET"),
    }
}