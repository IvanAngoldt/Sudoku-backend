package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	JWTSecret  string
	UsersURL   string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load(".env")

	// return &Config{
	// 	ServerPort: getEnv("AUTH_PORT", "8081"),
	// 	JWTSecret:  getEnv("JWT_SECRET", "twitch-prime"),
	// 	UsersURL:   getEnv("USERS_SERVICE_URL", "http://localhost:8082"),
	// }, nil

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT"),
		JWTSecret:  getEnv("JWT_SECRET"),
		UsersURL:   getEnv("USERS_SERVICE_URL"),
	}

	return cfg, nil
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("missing required environment variable: %s", key))
	}
	return val
}
