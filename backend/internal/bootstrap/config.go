package bootstrap

import (
	"errors"
	"os"
	"strconv"
)

// Config stores runtime configuration loaded from environment variables.
type Config struct {
	MySQLDSN  string
	RedisAddr string
	RedisPass string
	RedisDB   int
}

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
		MySQLDSN:  dsn,
		RedisAddr: redisAddr,
		RedisPass: redisPass,
		RedisDB:   redisDB,
	}, nil
}
