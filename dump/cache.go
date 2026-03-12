package dump

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// cachePath returns the platform-specific cache directory for a dump index.
// Uses os.UserCacheDir():
//
//	macOS: ~/Library/Caches/mcp-1c/<hash>
//	Linux: ~/.cache/mcp-1c/<hash>  (or $XDG_CACHE_HOME)
//	Windows: %LocalAppData%/mcp-1c/<hash>
func cachePath(dumpDir string) (string, error) {
	absDir, err := filepath.Abs(dumpDir)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256([]byte(absDir))
	hash := hex.EncodeToString(h[:8]) // first 16 hex chars

	cacheBase, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheBase, "mcp-1c", hash), nil
}

// cacheShardDirs returns sorted paths of shard_* subdirectories in cacheDir.
// Returns nil if the directory does not exist or contains no shards.
func cacheShardDirs(cacheDir string) []string {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "shard_") {
			dirs = append(dirs, filepath.Join(cacheDir, e.Name()))
		}
	}
	slices.Sort(dirs)
	return dirs
}
