package bootstrap

import (
	"errors"
	"os"
)

// Config stores runtime configuration loaded from environment variables.
type Config struct {
	MySQLDSN string
}

func LoadConfigFromEnv() (Config, error) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		return Config{}, errors.New("MYSQL_DSN is required")
	}

	return Config{MySQLDSN: dsn}, nil
}
