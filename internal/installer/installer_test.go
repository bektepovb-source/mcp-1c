package installer

import (
	"runtime"
	"strings"
	"testing"
)

func TestInstall_NonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this test only runs on non-Windows")
	}

	err := Install([]byte("test"), "/tmp/test-db")
	if err == nil {
		t.Fatal("expected error on non-Windows")
	}
	if !strings.Contains(err.Error(), "only supported on Windows") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFindPlatform_NonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this test only runs on non-Windows")
	}

	_, err := FindPlatform()
	if err == nil {
		t.Fatal("expected error on non-Windows")
	}
}
