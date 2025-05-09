package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	Port           string
	GameServiceURL string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	return &Config{
		DBHost:     os.Getenv("DBHost"),
		DBPort:     os.Getenv("DBPort"),
		DBUser:     os.Getenv("DBUser"),
		DBPassword: os.Getenv("DBPassword"),
		DBName:     os.Getenv("DBName"),

		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       0,

		Port:           os.Getenv("PORT"),
		GameServiceURL: os.Getenv("GAME_SERVICE_URL"),
	}, nil
}
