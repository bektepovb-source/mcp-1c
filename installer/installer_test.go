package installer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPlatformPatterns(t *testing.T) {
	patterns := platformPatterns()
	if len(patterns) == 0 {
		t.Fatalf("expected non-empty patterns for GOOS=%s", runtime.GOOS)
	}
	t.Logf("GOOS=%s, patterns: %v", runtime.GOOS, patterns)
}

func TestExtractXMLTag(t *testing.T) {
	xml := `<Properties>
		<CompatibilityMode>Version8_3_24</CompatibilityMode>
		<InterfaceCompatibilityMode>TaxiEnableVersion8_2</InterfaceCompatibilityMode>
	</Properties>`

	if got := extractXMLTag(xml, "CompatibilityMode"); got != "Version8_3_24" {
		t.Errorf("CompatibilityMode = %q, want Version8_3_24", got)
	}
	if got := extractXMLTag(xml, "InterfaceCompatibilityMode"); got != "TaxiEnableVersion8_2" {
		t.Errorf("InterfaceCompatibilityMode = %q, want TaxiEnableVersion8_2", got)
	}
	if got := extractXMLTag(xml, "Missing"); got != "" {
		t.Errorf("Missing = %q, want empty", got)
	}
}

func TestPatchExtensionXML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Configuration.xml")

	original := `<Properties>
			<ConfigurationExtensionCompatibilityMode>Version8_3_14</ConfigurationExtensionCompatibilityMode>
			<DefaultRunMode>ManagedApplication</DefaultRunMode>
		</Properties>`

	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := patchExtensionXML(path, "Version8_3_24", "TaxiEnableVersion8_2"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "<ConfigurationExtensionCompatibilityMode>Version8_3_24</ConfigurationExtensionCompatibilityMode>") {
		t.Error("CompatibilityMode not patched")
	}
	if !strings.Contains(content, "<InterfaceCompatibilityMode>TaxiEnableVersion8_2</InterfaceCompatibilityMode>") {
		t.Error("InterfaceCompatibilityMode not inserted")
	}
}

func TestPatchExtensionXML_InsertBoth(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Configuration.xml")

	original := `<Properties>
			<DefaultRunMode>ManagedApplication</DefaultRunMode>
		</Properties>`

	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := patchExtensionXML(path, "Version8_3_20", "Taxi"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "Version8_3_20") {
		t.Error("CompatibilityMode not inserted")
	}
	if !strings.Contains(content, "<InterfaceCompatibilityMode>Taxi</InterfaceCompatibilityMode>") {
		t.Error("InterfaceCompatibilityMode not inserted")
	}
}

func TestReplaceOrInsertXMLTag(t *testing.T) {
	// Replace existing tag.
	content := `<Foo>old</Foo>`
	got := replaceOrInsertXMLTag(content, "Foo", "new")
	if !strings.Contains(got, "<Foo>new</Foo>") {
		t.Errorf("replace failed: %s", got)
	}

	// Insert missing tag.
	content = `<Properties>
		</Properties>`
	got = replaceOrInsertXMLTag(content, "Bar", "val")
	if !strings.Contains(got, "<Bar>val</Bar>") {
		t.Errorf("insert failed: %s", got)
	}
}

func TestFindPlatform(t *testing.T) {
	path, err := FindPlatform()
	if err != nil {
		t.Logf("1C not installed (expected on CI): %v", err)
		return
	}
	t.Logf("Found 1C at: %s", path)
}
