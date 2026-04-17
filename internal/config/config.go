package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Port                    string
	MongoURI                string
	MongoDatabase           string
	JWTSecret               string
	SyncIntervalMinutes     int
	GeminiAPIKey            string
	SummaryCheckIntervalMin int
	SummaryBackfillDays     int
	CORSAllowLocalhost      bool
	AdminUIEnabled          bool
	AdminUser               string
	AdminPasswordHash       []byte // pre-hashed at Load time; nil if ADMIN_PASSWORD not set
	LogLevel                string
}

func Load() *Config {
	godotenv.Load()

	cfg := &Config{
		Port:                    getEnv("PORT", "8080"),
		MongoURI:                getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase:           getEnv("MONGODB_DATABASE", "conflux"),
		JWTSecret:               getEnv("JWT_SECRET", "change-me-in-production"),
		SyncIntervalMinutes:     getEnvInt("SYNC_INTERVAL_MINUTES", 15),
		GeminiAPIKey:            os.Getenv("GEMINI_API_KEY"),
		SummaryCheckIntervalMin: getEnvInt("SUMMARY_CHECK_INTERVAL_MIN", 30),
		SummaryBackfillDays:     getEnvInt("SUMMARY_BACKFILL_DAYS", 7),
		CORSAllowLocalhost:      getEnvBool("CORS_ALLOW_LOCALHOST", false),
		AdminUIEnabled:          getEnvBool("ADMIN_UI_ENABLED", true),
		AdminUser:               os.Getenv("ADMIN_USER"),
		LogLevel:                getEnv("LOG_LEVEL", "info"),
	}

	// Hash ADMIN_PASSWORD once at startup; raw value never persists in memory.
	if rawPwd := os.Getenv("ADMIN_PASSWORD"); rawPwd != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(rawPwd), 12)
		if err != nil {
			log.Fatal().Err(err).Msg("[config.Load] failed to hash ADMIN_PASSWORD")
		}
		cfg.AdminPasswordHash = hash
		// Best-effort: remove from environment so child processes / crash dumps don't leak it.
		_ = os.Unsetenv("ADMIN_PASSWORD")
	}

	return cfg
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
