package config

import "os"

// Config holds the MCP server configuration.
type Config struct {
	BaseURL  string
	User     string
	Password string
}

// Load reads configuration from environment variables.
// CLI flags should be parsed in main.go before calling Load.
// Environment variables override any values already set in the Config.
func Load() *Config {
	cfg := &Config{
		BaseURL: "http://localhost:8080/hs/mcp",
	}

	if v := os.Getenv("MCP_1C_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("MCP_1C_USER"); v != "" {
		cfg.User = v
	}
	if v := os.Getenv("MCP_1C_PASSWORD"); v != "" {
		cfg.Password = v
	}

	return cfg
}
