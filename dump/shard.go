package dump

import (
	"fmt"
	"hash/fnv"
	"os"
	"runtime"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

// shardCount returns the optimal number of index shards for the given file count.
// Uses 1 shard per 2000 files, clamped to [1, runtime.NumCPU()].
func shardCount(totalFiles int) int {
	if totalFiles <= 0 {
		return 1
	}
	n := totalFiles / 2000
	return max(1, min(runtime.NumCPU(), n))
}

// shardForID returns a deterministic shard index for the given document ID
// using FNV-32a hash. Result is in [0, n).
func shardForID(id string, n int) int {
	if n <= 1 {
		return 0
	}
	h := fnv.New32a()
	h.Write([]byte(id))
	return int(h.Sum32() % uint32(n))
}

// splitByHash distributes items across n groups using shardForID.
// Each item lands in exactly one group determined by its hash.
func splitByHash(items []string, n int) [][]string {
	groups := make([][]string, n)
	for _, item := range items {
		i := shardForID(item, n)
		groups[i] = append(groups[i], item)
	}
	return groups
}

// buildShard creates a single Bleve shard index at the given path.
// It indexes the provided names using content from contentByName.
// shardID and totalShards are used for progress reporting.
// The caller must supply a pre-built bslMapping to avoid rebuilding it per shard.
func buildShard(path string, names []string, contentByName map[string]string, shardID, totalShards int, bslMapping *mapping.IndexMappingImpl) (bleve.Index, error) {
	blevIdx, err := bleve.NewUsing(path, bslMapping, "scorch", "scorch", map[string]any{
		"unsafe_batch": true,
	})
	if err != nil {
		return nil, fmt.Errorf("creating bleve shard %d: %w", shardID, err)
	}

	total := len(names)
	const batchSize = 5000
	step := max(total/10, 100)

	batch := blevIdx.NewBatch()
	for i, name := range names {
		parts := parseModuleName(name)
		doc := bslDocument{
			Name:     parts.name,
			Category: parts.category,
			Module:   parts.module,
			Content:  contentByName[name],
		}
		batch.Index(name, doc)

		if (i+1)%batchSize == 0 || i+1 == total {
			if err := blevIdx.Batch(batch); err != nil {
				blevIdx.Close()
				return nil, fmt.Errorf("shard %d batch: %w", shardID, err)
			}
			batch = blevIdx.NewBatch()
		}

		if (i+1)%step == 0 || i+1 == total {
			fmt.Fprintf(os.Stderr, "\rShard %d/%d: indexed %d/%d modules", shardID+1, totalShards, i+1, total)
		}
	}
	if total > 0 {
		fmt.Fprintln(os.Stderr)
	}

	return blevIdx, nil
}
