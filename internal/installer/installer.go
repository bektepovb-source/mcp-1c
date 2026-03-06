package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Install writes the .cfe to a temp file and installs it into the 1C database.
func Install(cfeData []byte, dbPath string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-install is only supported on Windows")
	}

	platformExe, err := FindPlatform()
	if err != nil {
		return fmt.Errorf("finding 1C platform: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "mcp-1c-*.cfe")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(cfeData); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing cfe: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command(platformExe,
		"DESIGNER",
		"/F", dbPath,
		"/LoadCfg", tmpFile.Name(),
		"/Extension", "MCP_HTTPService",
		"/UpdateDBCfg",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("1C DESIGNER failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// FindPlatform searches for the 1C platform executable.
func FindPlatform() (string, error) {
	patterns := []string{
		`C:\Program Files\1cv8\8.*\bin\1cv8.exe`,
		`C:\Program Files (x86)\1cv8\8.*\bin\1cv8.exe`,
		`C:\Program Files\1cv8t\8.*\bin\1cv8t.exe`,
		`C:\Program Files (x86)\1cv8t\8.*\bin\1cv8t.exe`,
	}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			return matches[len(matches)-1], nil // latest version
		}
	}

	return "", fmt.Errorf("1C platform not found in standard paths")
}
