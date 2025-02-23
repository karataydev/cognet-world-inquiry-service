package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisAddress  string
	RedisPassword string
	ServerPort    string
}

var AppConfig Config

func Load() error {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}

	AppConfig = Config{
		RedisAddress:  os.Getenv("REDIS_ADDRESS"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		ServerPort:    os.Getenv("SERVER_PORT"),
	}
	fmt.Println("Configuration loaded successfully")
	return nil
}
