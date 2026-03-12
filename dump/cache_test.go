package dump

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCachePath_NonEmpty(t *testing.T) {
	p, err := cachePath("/some/dump/dir")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	if p == "" {
		t.Fatal("expected non-empty path")
	}
}

func TestCachePath_Deterministic(t *testing.T) {
	p1, err := cachePath("/some/dump/dir")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	p2, err := cachePath("/some/dump/dir")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	if p1 != p2 {
		t.Errorf("expected deterministic result, got %q and %q", p1, p2)
	}
}

func TestCachePath_DifferentDirs(t *testing.T) {
	p1, err := cachePath("/dir/one")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	p2, err := cachePath("/dir/two")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	if p1 == p2 {
		t.Errorf("expected different paths for different dirs, got same: %q", p1)
	}
}

func TestCachePath_ContainsMcp1c(t *testing.T) {
	p, err := cachePath("/any/dir")
	if err != nil {
		t.Fatalf("cachePath: %v", err)
	}
	if !strings.Contains(p, "mcp-1c") {
		t.Errorf("expected path to contain 'mcp-1c', got %q", p)
	}
}

func TestCacheShardDirs(t *testing.T) {
	t.Run("returns sorted shard dirs", func(t *testing.T) {
		dir := t.TempDir()
		// Create shard dirs in non-sorted order.
		for _, name := range []string{"shard_2", "shard_0", "shard_1"} {
			if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		// Create a non-shard dir and a file that should be ignored.
		os.Mkdir(filepath.Join(dir, "other"), 0o755)
		os.WriteFile(filepath.Join(dir, "shard_file.txt"), []byte("x"), 0o644)

		dirs := cacheShardDirs(dir)
		if len(dirs) != 3 {
			t.Fatalf("expected 3 shard dirs, got %d: %v", len(dirs), dirs)
		}
		for i, want := range []string{"shard_0", "shard_1", "shard_2"} {
			expected := filepath.Join(dir, want)
			if dirs[i] != expected {
				t.Errorf("dirs[%d] = %q, want %q", i, dirs[i], expected)
			}
		}
	})

	t.Run("non-existent dir returns nil", func(t *testing.T) {
		dirs := cacheShardDirs("/nonexistent/path/that/does/not/exist")
		if dirs != nil {
			t.Errorf("expected nil for non-existent dir, got %v", dirs)
		}
	})

	t.Run("empty dir returns nil", func(t *testing.T) {
		dir := t.TempDir()
		dirs := cacheShardDirs(dir)
		if dirs != nil {
			t.Errorf("expected nil for empty dir, got %v", dirs)
		}
	})
}
