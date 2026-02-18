package bootstrap

import (
	"errors"
	"os"
	"strconv"
)

// Config stores runtime configuration loaded from environment variables.
type Config struct {
	MySQLDSN          string
	RedisAddr         string
	RedisPass         string
	RedisDB           int
	PublicBaseURL     string
	MediaCacheDir     string
	AdminJWTSecret    string
	AdminUsername     string
	AdminPasswordHash string
	SummaryProvider   string
	SummaryAPIBase    string
	SummaryAPIKey     string
}

const (
	defaultAdminJWTSecret    = "dev-admin-jwt-secret-change-me"
	defaultAdminUsername     = "admin"
	defaultAdminPasswordHash = "$2a$10$8VE/OERwmsxYhBXnYs2ULuDx.Zw78wZMnXPIovrY8SQEpKdYmgNKK" // admin123456
	defaultPublicBaseURL     = "http://localhost:8080"
	defaultMediaCacheDir     = ".worktrees/media-cache"
	defaultSummaryProvider   = "openai"
	defaultSummaryAPIBase    = "https://api.openai.com/v1"
)

func LoadConfigFromEnv() (Config, error) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		return Config{}, errors.New("MYSQL_DSN is required")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return Config{}, errors.New("REDIS_ADDR is required")
	}

	redisPass := os.Getenv("REDIS_PASSWORD")
	redisDB := 0
	if raw := os.Getenv("REDIS_DB"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, errors.New("REDIS_DB must be a number")
		}
		redisDB = value
	}

	return Config{
		MySQLDSN:          dsn,
		RedisAddr:         redisAddr,
		RedisPass:         redisPass,
		RedisDB:           redisDB,
		PublicBaseURL:     getenvOrDefault("PUBLIC_BASE_URL", defaultPublicBaseURL),
		MediaCacheDir:     getenvOrDefault("MEDIA_CACHE_DIR", defaultMediaCacheDir),
		AdminJWTSecret:    getenvOrDefault("ADMIN_JWT_SECRET", defaultAdminJWTSecret),
		AdminUsername:     getenvOrDefault("ADMIN_USERNAME", defaultAdminUsername),
		AdminPasswordHash: getenvOrDefault("ADMIN_PASSWORD_HASH", defaultAdminPasswordHash),
		SummaryProvider:   getenvOrDefault("SUMMARY_PROVIDER", defaultSummaryProvider),
		SummaryAPIBase:    getenvOrDefault("SUMMARY_API_BASE", defaultSummaryAPIBase),
		SummaryAPIKey:     os.Getenv("SUMMARY_API_KEY"),
	}, nil
}

func getenvOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
