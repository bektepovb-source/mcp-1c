package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Unset env vars to test defaults.
	t.Setenv("MCP_1C_BASE_URL", "")
	t.Setenv("MCP_1C_USER", "")
	t.Setenv("MCP_1C_PASSWORD", "")

	cfg := Load()

	if cfg.BaseURL != "http://localhost:8080/hs/mcp" {
		t.Errorf("expected default base URL, got %s", cfg.BaseURL)
	}
	if cfg.User != "" {
		t.Errorf("expected empty user, got %s", cfg.User)
	}
	if cfg.Password != "" {
		t.Errorf("expected empty password, got %s", cfg.Password)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("MCP_1C_BASE_URL", "http://custom:9090/api")
	t.Setenv("MCP_1C_USER", "admin")
	t.Setenv("MCP_1C_PASSWORD", "secret")

	cfg := Load()

	if cfg.BaseURL != "http://custom:9090/api" {
		t.Errorf("expected overridden base URL, got %s", cfg.BaseURL)
	}
	if cfg.User != "admin" {
		t.Errorf("expected user admin, got %s", cfg.User)
	}
	if cfg.Password != "secret" {
		t.Errorf("expected password secret, got %s", cfg.Password)
	}
}

func TestLoadPartialOverride(t *testing.T) {
	t.Setenv("MCP_1C_BASE_URL", "")
	t.Setenv("MCP_1C_USER", "operator")
	t.Setenv("MCP_1C_PASSWORD", "")

	cfg := Load()

	if cfg.BaseURL != "http://localhost:8080/hs/mcp" {
		t.Errorf("expected default base URL, got %s", cfg.BaseURL)
	}
	if cfg.User != "operator" {
		t.Errorf("expected user operator, got %s", cfg.User)
	}
	if cfg.Password != "" {
		t.Errorf("expected empty password, got %s", cfg.Password)
	}
}
