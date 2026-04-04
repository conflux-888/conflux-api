package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	MongoURI            string
	MongoDatabase       string
	JWTSecret           string
	ACLEDUsername        string
	ACLEDPassword        string
	SyncIntervalMinutes int
	LogLevel            string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:                getEnv("PORT", "8080"),
		MongoURI:            getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase:       getEnv("MONGODB_DATABASE", "conflux"),
		JWTSecret:           getEnv("JWT_SECRET", "change-me-in-production"),
		ACLEDUsername:        os.Getenv("ACLED_USERNAME"),
		ACLEDPassword:        os.Getenv("ACLED_PASSWORD"),
		SyncIntervalMinutes: getEnvInt("SYNC_INTERVAL_MINUTES", 15),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
