package dump

import (
	"os"
	"testing"
)

// realDumpDir points to a real 1C dump for benchmarking.
// Benchmarks are skipped if this directory does not exist.
const realDumpDir = "/Users/igoroot/GolandProjects/mcp/dumps/dump_2"

// loadTestModules reads all BSL files from the real dump directory into memory.
// Returns names and contentByName for use in build benchmarks.
// Calls b.Skip if the dump directory is missing.
func loadTestModules(b *testing.B) ([]string, map[string]string) {
	b.Helper()

	if _, err := os.Stat(realDumpDir); os.IsNotExist(err) {
		b.Skipf("dump directory %s does not exist, skipping benchmark", realDumpDir)
	}

	idx := &Index{
		dir:           realDumpDir,
		contentByName: make(map[string]string),
	}
	if err := idx.loadBSLFiles(realDumpDir); err != nil {
		b.Fatalf("loadBSLFiles: %v", err)
	}
	b.Logf("Loaded %d BSL modules from %s", len(idx.names), realDumpDir)

	return idx.names, idx.contentByName
}

// BenchmarkBuildIndex_Batch measures the current NewUsing + manual batch approach.
func BenchmarkBuildIndex_Batch(b *testing.B) {
	names, contentByName := loadTestModules(b)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		dir := b.TempDir()
		indexPath := dir + "/index"

		blevIdx, err := buildIndexBatch(indexPath, names, contentByName)
		if err != nil {
			b.Fatalf("buildIndexBatch: %v", err)
		}
		blevIdx.Close()
	}
}

// BenchmarkBuildIndex_Builder measures the offline NewBuilder approach.
func BenchmarkBuildIndex_Builder(b *testing.B) {
	names, contentByName := loadTestModules(b)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		dir := b.TempDir()
		indexPath := dir + "/index"

		blevIdx, err := buildIndexBuilder(indexPath, names, contentByName)
		if err != nil {
			b.Fatalf("buildIndexBuilder: %v", err)
		}
		blevIdx.Close()
	}
}

// openRealIndex builds a fresh index from the real dump for search benchmarks.
func openRealIndex(b *testing.B) *Index {
	b.Helper()

	if _, err := os.Stat(realDumpDir); os.IsNotExist(err) {
		b.Skipf("dump directory %s does not exist, skipping benchmark", realDumpDir)
	}

	idx, err := NewIndex(realDumpDir, true)
	if err != nil {
		b.Fatalf("NewIndex: %v", err)
	}
	b.Logf("Index built: %d modules", idx.ModuleCount())

	return idx
}

// BenchmarkSearch_Smart measures BM25 full-text search performance.
func BenchmarkSearch_Smart(b *testing.B) {
	idx := openRealIndex(b)
	defer idx.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _, err := idx.Search(SearchParams{
			Query: "Процедура ПередЗаписью",
			Mode:  SearchModeSmart,
			Limit: 50,
		})
		if err != nil {
			b.Fatalf("Search smart: %v", err)
		}
	}
}

// BenchmarkSearch_Regex measures regex search performance.
func BenchmarkSearch_Regex(b *testing.B) {
	idx := openRealIndex(b)
	defer idx.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _, err := idx.Search(SearchParams{
			Query: `Процедура\s+\w+Записью`,
			Mode:  SearchModeRegex,
			Limit: 50,
		})
		if err != nil {
			b.Fatalf("Search regex: %v", err)
		}
	}
}

// BenchmarkSearch_Exact measures exact (case-insensitive substring) search performance.
func BenchmarkSearch_Exact(b *testing.B) {
	idx := openRealIndex(b)
	defer idx.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _, err := idx.Search(SearchParams{
			Query: "ОбработкаПроведения",
			Mode:  SearchModeExact,
			Limit: 50,
		})
		if err != nil {
			b.Fatalf("Search exact: %v", err)
		}
	}
}
