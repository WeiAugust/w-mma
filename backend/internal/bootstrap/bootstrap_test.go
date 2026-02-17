package bootstrap

import "testing"

func TestLoadConfig_MissingEnvFails(t *testing.T) {
	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected error when required env is missing")
	}
}
