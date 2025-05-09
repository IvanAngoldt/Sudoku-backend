package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	JWTSecret  string
	UsersURL   string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		ServerPort: getEnv("AUTH_PORT", "8081"),
		JWTSecret:  getEnv("JWT_SECRET", "twitch-prime"),
		UsersURL:   getEnv("USERS_SERVICE_URL", "http://localhost:8082"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
