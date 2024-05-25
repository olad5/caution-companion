package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Configurations struct {
	DatabaseUrl  string
	DatabaseName string
	Port         string
	JwtSecretKey string
	CacheAddress string
	LogLevel     string
}

func GetConfig(filepath string) *Configurations {
	err := godotenv.Load(filepath)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configurations := Configurations{
		DatabaseUrl:  os.Getenv("DATABASE_URL"),
		DatabaseName: os.Getenv("DATABASE_NAME"),
		Port:         os.Getenv("PORT"),
		JwtSecretKey: os.Getenv("SECRET_KEY"),
		CacheAddress: os.Getenv("REDIS_URL"),
		LogLevel:     os.Getenv("LOG_LEVEL"),
	}

	return &configurations
}
