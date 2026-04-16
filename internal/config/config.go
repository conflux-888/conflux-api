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
	SyncIntervalMinutes     int
	GeminiAPIKey            string
	SummaryCheckIntervalMin int
	SummaryBackfillDays     int
	CORSAllowLocalhost      bool
	AdminUIEnabled          bool
	LogLevel                string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:                getEnv("PORT", "8080"),
		MongoURI:            getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase:       getEnv("MONGODB_DATABASE", "conflux"),
		JWTSecret:           getEnv("JWT_SECRET", "change-me-in-production"),
		SyncIntervalMinutes:     getEnvInt("SYNC_INTERVAL_MINUTES", 15),
		GeminiAPIKey:            os.Getenv("GEMINI_API_KEY"),
		SummaryCheckIntervalMin: getEnvInt("SUMMARY_CHECK_INTERVAL_MIN", 30),
		SummaryBackfillDays:     getEnvInt("SUMMARY_BACKFILL_DAYS", 7),
		CORSAllowLocalhost:      getEnvBool("CORS_ALLOW_LOCALHOST", false),
		AdminUIEnabled:          getEnvBool("ADMIN_UI_ENABLED", true),
		LogLevel:                getEnv("LOG_LEVEL", "info"),
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

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
