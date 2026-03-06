package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/feenlace/mcp-1c/internal/config"
	"github.com/feenlace/mcp-1c/internal/installer"
	"github.com/feenlace/mcp-1c/internal/onec"
	"github.com/feenlace/mcp-1c/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed extension/MCP_HTTPService.cfe
var embeddedCFE []byte

const expectedExtensionVersion = "0.2.0"

func main() {
	baseURL := flag.String("base", "", "Base URL of 1C HTTP service")
	user := flag.String("user", "", "1C HTTP service user")
	password := flag.String("password", "", "1C HTTP service password")
	installDB := flag.String("install", "", "Install extension into 1C database at given path")
	flag.Parse()

	// Install mode.
	if *installDB != "" {
		fmt.Println("Installing MCP extension into 1C database...")
		if err := installer.Install(embeddedCFE, *installDB); err != nil {
			fmt.Fprintf(os.Stderr, "Installation error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Extension installed successfully.")
		return
	}

	// Load defaults and env var overrides.
	cfg := config.Load()

	// CLI flags take highest priority (override env vars).
	if *baseURL != "" {
		cfg.BaseURL = *baseURL
	}
	if *user != "" {
		cfg.User = *user
	}
	if *password != "" {
		cfg.Password = *password
	}

	client := onec.NewClient(cfg.BaseURL, cfg.User, cfg.Password)

	// Version check (non-blocking).
	checkExtensionVersion(client)

	s := server.New(client)

	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "mcp-1c error: %v\n", err)
		os.Exit(1)
	}
}

func checkExtensionVersion(client *onec.Client) {
	var ver onec.VersionInfo
	if err := client.Get(context.Background(), "/version", &ver); err != nil {
		// Version endpoint may not exist in older extensions — skip silently.
		return
	}
	if ver.Version != expectedExtensionVersion {
		fmt.Fprintf(os.Stderr, "WARNING: Extension version %s, expected %s. Update: mcp-1c --install \"path\\to\\db\"\n",
			ver.Version, expectedExtensionVersion)
	}
}
