package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Configurations struct {
	DatabaseUrl             string
	DatabaseName            string
	Port                    string
	JwtSecretKey            string
	AppName                 string
	CacheAddress            string
	AuthSessionTTLInMinutes int
	LogLevel                string
	Environment             string
	CloudinaryUrl           string
}

func GetConfig(filepath string) *Configurations {
	err := godotenv.Load(filepath)
	environment := os.Getenv("ENVIRONMENT")

	if err != nil && environment != "production" {
		log.Fatal("Error loading .env file")
	}

	authSessionTTLInMinutes, err := strconv.Atoi(os.Getenv("AUTH_SESSION_TTL"))
	if err != nil {
		log.Fatal("Error loading AUTH_SESSION_TTL from .env file")
	}

	configurations := Configurations{
		DatabaseUrl:             os.Getenv("DATABASE_URL"),
		DatabaseName:            os.Getenv("DATABASE_NAME"),
		Port:                    os.Getenv("PORT"),
		JwtSecretKey:            os.Getenv("SECRET_KEY"),
		CacheAddress:            os.Getenv("REDIS_URL"),
		LogLevel:                os.Getenv("LOG_LEVEL"),
		AppName:                 os.Getenv("APP_NAME"),
		CloudinaryUrl:           os.Getenv("CLOUDINARY_URL"),
		AuthSessionTTLInMinutes: authSessionTTLInMinutes,
		Environment:             environment,
	}

	return &configurations
}
