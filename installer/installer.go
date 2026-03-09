package installer

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
        "regexp"
	"runtime"
	"strings"
)

const extensionName = "MCP_HTTPService"


// Install extracts embedded XML sources to a temp dir, reads the target database's
// compatibility settings, patches the extension XML to match, and loads it into 1C.
// If platformExe is empty, the platform is auto-detected.
func Install(srcFS embed.FS, dbPath, platformExe, dbUser, dbPassword string) error {
	if platformExe == "" {
		var err error
		platformExe, err = FindPlatform()
		if err != nil {
			return fmt.Errorf("finding 1C platform: %w", err)
		}
	}
	fmt.Printf("Platform: %s\n", platformExe)

	// Extract extension XML sources to temp dir.
	extDir, err := os.MkdirTemp("", "mcp-1c-ext-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(extDir)

	if err := extractFS(srcFS, "src", extDir); err != nil {
		return fmt.Errorf("extracting extension sources: %w", err)
	}

	// TODO: auto-detect compatibility mode from database and patch extension XML.
	// Disabled for now — hardcoded Version8_3_14 in extension XML works for most databases.
	//
	// dumpDir, err := os.MkdirTemp("", "mcp-1c-dump-*")
	// defer os.RemoveAll(dumpDir)
	// listFile := filepath.Join(dumpDir, "dump-list.txt")
	// os.WriteFile(listFile, []byte("Configuration\n"), 0o644)
	// runDesigner(platformExe, dbPath, dbUser, dbPassword, "/DumpConfigToFiles", dumpDir, "-listFile", listFile)
	// mainCfg, _ := os.ReadFile(filepath.Join(dumpDir, "Configuration.xml"))
	// compatMode := extractXMLTag(string(mainCfg), "CompatibilityMode")
	// interfaceMode := extractXMLTag(string(mainCfg), "InterfaceCompatibilityMode")
	// patchExtensionXML(filepath.Join(extDir, "Configuration.xml"), compatMode, interfaceMode)

	// Load extension XML into extension configuration.
	fmt.Println("Loading extension into database...")
	if err := runDesigner(platformExe, dbPath, dbUser, dbPassword,
		"/LoadConfigFromFiles", extDir,
		"-Extension", extensionName,
	); err != nil {
		return fmt.Errorf("loading extension config: %w", err)
	}

	// Apply extension to the database (separate call required).
	fmt.Println("Updating database...")
	return runDesigner(platformExe, dbPath, dbUser, dbPassword,
		"/UpdateDBCfg",
		"-Extension", extensionName,
	)
}

// runDesigner executes 1C DESIGNER with given arguments, capturing output via /Out.
func runDesigner(platformExe, dbPath, dbUser, dbPassword string, extraArgs ...string) error {
	logFile, err := os.CreateTemp("", "mcp-1c-log-*.txt")
	if err != nil {
		return fmt.Errorf("creating log file: %w", err)
	}
	logFile.Close()
	defer os.Remove(logFile.Name())

	args := []string{"DESIGNER", "/F", dbPath}
	if dbUser != "" {
		args = append(args, "/N", dbUser)
	}
	if dbPassword != "" {
		args = append(args, "/P", dbPassword)
	}
	args = append(args, extraArgs...)
	args = append(args, "/Out", logFile.Name(), "/DisableStartupDialogs", "/DisableStartupMessages")

	cmd := exec.Command(platformExe, args...)
	cmd.CombinedOutput() //nolint:errcheck // exit code checked via log
	logData, _ := os.ReadFile(logFile.Name())
	logStr := strings.TrimSpace(string(bytes.TrimLeft(logData, "\xef\xbb\xbf")))

	if cmd.ProcessState == nil {
		return fmt.Errorf("1C DESIGNER failed to start: %s", platformExe)
	}
	if !cmd.ProcessState.Success() {
		if logStr != "" {
			return fmt.Errorf("1C DESIGNER failed (exit code %d):\n%s", cmd.ProcessState.ExitCode(), logStr)
		}
		return fmt.Errorf("1C DESIGNER failed with exit code %d (no log output)", cmd.ProcessState.ExitCode())
	}
	if logStr != "" {
		fmt.Println(logStr)
	}
	return nil
}

// extractXMLTag extracts the text content of a simple XML tag like <TagName>value</TagName>.
func extractXMLTag(xml, tag string) string {
	re := regexp.MustCompile(`<` + tag + `>([^<]+)</` + tag + `>`)
	m := re.FindStringSubmatch(xml)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// replaceOrInsertXMLTag replaces an existing XML tag value or inserts a new tag before </Properties>.
func replaceOrInsertXMLTag(content, tagName, value string) string {
	re := regexp.MustCompile(`<` + tagName + `>[^<]+</` + tagName + `>`)
	replacement := "<" + tagName + ">" + value + "</" + tagName + ">"
	if re.MatchString(content) {
		return re.ReplaceAllString(content, replacement)
	}
	return strings.Replace(content, "</Properties>",
		"\t\t\t"+replacement+"\n\t\t</Properties>", 1)
}

// patchExtensionXML updates ConfigurationExtensionCompatibilityMode and InterfaceCompatibilityMode
// in the extension's Configuration.xml to match the target database.
func patchExtensionXML(path, compatMode, interfaceMode string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	if compatMode != "" {
		content = replaceOrInsertXMLTag(content, "ConfigurationExtensionCompatibilityMode", compatMode)
	}
	if interfaceMode != "" {
		content = replaceOrInsertXMLTag(content, "InterfaceCompatibilityMode", interfaceMode)
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

// extractFS copies files from an embed.FS subtree into a directory on disk.
func extractFS(fsys embed.FS, root, destDir string) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		target := filepath.Join(destDir, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fsys.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// FindPlatform searches for the 1C platform executable on the current OS.
// Returns the last match from sorted glob results (latest version by lexical order).
func FindPlatform() (string, error) {
	patterns := platformPatterns()
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			return matches[len(matches)-1], nil
		}
	}
	return "", fmt.Errorf("1C platform not found in standard paths")
}

// platformPatterns returns glob patterns for finding 1C platform binary on the current OS.
func platformPatterns() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			`C:\Program Files\1cv8\8.*\bin\1cv8.exe`,
			`C:\Program Files (x86)\1cv8\8.*\bin\1cv8.exe`,
			`C:\Program Files\1cv8t\8.*\bin\1cv8t.exe`,
			`C:\Program Files (x86)\1cv8t\8.*\bin\1cv8t.exe`,
			`C:\Program Files\1cv82\8.*\bin\1cv8.exe`,
			`C:\Program Files (x86)\1cv82\8.*\bin\1cv8.exe`,
		}
	case "darwin":
		return []string{
			"/Applications/1cv8.localized/*/1cv8.app/Contents/MacOS/1cv8",
			"/Applications/1cv8t.localized/*/1cv8t.app/Contents/MacOS/1cv8t",
		}
	case "linux":
		return []string{
			"/opt/1cv8/x86_64/8.3.*/1cv8",
			"/opt/1cv8/x86_64/8.5.*/1cv8",
			"/opt/1C/v8.3/x86_64/1cv8",
		}
	default:
		return nil
	}
}
