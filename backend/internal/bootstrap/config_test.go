package bootstrap

import (
	"os"
	"testing"
)

func TestLoadConfig_RequiresMySQLAndRedis(t *testing.T) {
	t.Setenv("MYSQL_DSN", "")
	t.Setenv("REDIS_ADDR", "")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected error when required env missing")
	}
}

func TestLoadConfig_LoadsRedisSettings(t *testing.T) {
	t.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/bajiaozhi")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("REDIS_PASSWORD", "secret")
	t.Setenv("REDIS_DB", "3")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RedisAddr != "localhost:6379" {
		t.Fatalf("unexpected redis addr: %s", cfg.RedisAddr)
	}
	if cfg.RedisPass != "secret" {
		t.Fatalf("unexpected redis password")
	}
	if cfg.RedisDB != 3 {
		t.Fatalf("unexpected redis db: %d", cfg.RedisDB)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
