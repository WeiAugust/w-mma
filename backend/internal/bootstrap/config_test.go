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
	t.Setenv("ADMIN_JWT_SECRET", "jwt-secret")
	t.Setenv("ADMIN_USERNAME", "ops-admin")
	t.Setenv("ADMIN_PASSWORD_HASH", "$2a$10$dummy")
	t.Setenv("SUMMARY_PROVIDER", "openai")
	t.Setenv("SUMMARY_API_BASE", "https://api.example.com/v1")
	t.Setenv("SUMMARY_API_KEY", "sk-123")

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
	if cfg.AdminJWTSecret != "jwt-secret" {
		t.Fatalf("unexpected admin jwt secret")
	}
	if cfg.AdminUsername != "ops-admin" {
		t.Fatalf("unexpected admin username")
	}
	if cfg.AdminPasswordHash != "$2a$10$dummy" {
		t.Fatalf("unexpected admin password hash")
	}
	if cfg.SummaryProvider != "openai" {
		t.Fatalf("unexpected summary provider")
	}
	if cfg.SummaryAPIBase != "https://api.example.com/v1" {
		t.Fatalf("unexpected summary api base")
	}
	if cfg.SummaryAPIKey != "sk-123" {
		t.Fatalf("unexpected summary api key")
	}
}

func TestLoadConfig_UsesAdminDefaultsWhenUnset(t *testing.T) {
	t.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/bajiaozhi")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("ADMIN_JWT_SECRET", "")
	t.Setenv("ADMIN_USERNAME", "")
	t.Setenv("ADMIN_PASSWORD_HASH", "")
	t.Setenv("SUMMARY_PROVIDER", "")
	t.Setenv("SUMMARY_API_BASE", "")
	t.Setenv("SUMMARY_API_KEY", "")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AdminJWTSecret != defaultAdminJWTSecret {
		t.Fatalf("expected default jwt secret")
	}
	if cfg.AdminUsername != defaultAdminUsername {
		t.Fatalf("expected default admin username")
	}
	if cfg.AdminPasswordHash != defaultAdminPasswordHash {
		t.Fatalf("expected default admin password hash")
	}
	if cfg.SummaryProvider != defaultSummaryProvider {
		t.Fatalf("expected default summary provider")
	}
	if cfg.SummaryAPIBase != defaultSummaryAPIBase {
		t.Fatalf("expected default summary api base")
	}
	if cfg.SummaryAPIKey != "" {
		t.Fatalf("expected empty summary api key")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
