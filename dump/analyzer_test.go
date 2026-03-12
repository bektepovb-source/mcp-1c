package dump

import (
	"testing"

	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/feenlace/mcp-1c/bsl"
)

func TestBuildSynonymMap(t *testing.T) {
	m := buildSynonymMap()

	// Keyword pairs are bidirectional.
	tests := []struct {
		key, want string
	}{
		{"процедура", "procedure"},
		{"procedure", "процедура"},
		{"функция", "function"},
		{"function", "функция"},
		{"если", "if"},
		{"if", "если"},
		{"возврат", "return"},
		{"return", "возврат"},
		{"истина", "true"},
		{"true", "истина"},
	}
	for _, tt := range tests {
		got, ok := m[tt.key]
		if !ok {
			t.Errorf("synonym map missing key %q", tt.key)
			continue
		}
		if got != tt.want {
			t.Errorf("synonym[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}

	// Built-in function pairs (from bsl.BuiltinFunctions).
	fnTests := []struct {
		key, want string
	}{
		{"стрнайти", "strfind"},
		{"strfind", "стрнайти"},
		{"стрдлина", "strlen"},
		{"strlen", "стрдлина"},
		{"текущаядата", "currentdate"},
		{"currentdate", "текущаядата"},
	}
	for _, tt := range fnTests {
		got, ok := m[tt.key]
		if !ok {
			t.Errorf("synonym map missing built-in function key %q", tt.key)
			continue
		}
		if got != tt.want {
			t.Errorf("synonym[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}

	// Map should have at least as many entries as built-in function pairs * 2
	// (bidirectional) plus keyword pairs. Use bsl.BuiltinFunctions as a baseline.
	minExpected := len(bsl.BuiltinFunctions) * 2
	if len(m) < minExpected {
		t.Errorf("expected at least %d entries (from BuiltinFunctions), got %d", minExpected, len(m))
	}
}

func TestBSLSynonymFilter(t *testing.T) {
	f := newBSLSynonymFilter()

	input := analysis.TokenStream{
		{Term: []byte("стрнайти"), Position: 1, Start: 0, End: 16},
	}

	output := f.Filter(input)

	if len(output) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(output))
	}

	// Original token preserved.
	if string(output[0].Term) != "стрнайти" {
		t.Errorf("token[0] = %q, want %q", output[0].Term, "стрнайти")
	}
	if output[0].Position != 1 {
		t.Errorf("token[0].Position = %d, want 1", output[0].Position)
	}

	// Synonym injected at same position.
	if string(output[1].Term) != "strfind" {
		t.Errorf("token[1] = %q, want %q", output[1].Term, "strfind")
	}
	if output[1].Position != 1 {
		t.Errorf("token[1].Position = %d, want 1 (same as original)", output[1].Position)
	}
}

func TestBSLSynonymFilter_Reverse(t *testing.T) {
	f := newBSLSynonymFilter()

	input := analysis.TokenStream{
		{Term: []byte("procedure"), Position: 1, Start: 0, End: 9},
	}

	output := f.Filter(input)

	if len(output) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(output))
	}
	if string(output[0].Term) != "procedure" {
		t.Errorf("token[0] = %q, want %q", output[0].Term, "procedure")
	}
	if string(output[1].Term) != "процедура" {
		t.Errorf("token[1] = %q, want %q", output[1].Term, "процедура")
	}
}

func TestBSLSynonymFilter_NoMatch(t *testing.T) {
	f := newBSLSynonymFilter()

	input := analysis.TokenStream{
		{Term: []byte("myvar"), Position: 1, Start: 0, End: 5},
	}

	output := f.Filter(input)

	if len(output) != 1 {
		t.Fatalf("expected 1 token (pass-through), got %d", len(output))
	}
	if string(output[0].Term) != "myvar" {
		t.Errorf("token[0] = %q, want %q", output[0].Term, "myvar")
	}
}

func TestBSLSynonymFilter_MultipleTokens(t *testing.T) {
	f := newBSLSynonymFilter()

	input := analysis.TokenStream{
		{Term: []byte("процедура"), Position: 1, Start: 0, End: 18},
		{Term: []byte("обновить"), Position: 2, Start: 19, End: 35},
		{Term: []byte("если"), Position: 3, Start: 36, End: 44},
	}

	output := f.Filter(input)

	// "процедура" -> + "procedure", "обновить" -> no synonym, "если" -> + "if"
	if len(output) != 5 {
		t.Fatalf("expected 5 tokens, got %d", len(output))
	}

	// Check positions: synonyms share position with original.
	if output[0].Position != 1 || string(output[0].Term) != "процедура" {
		t.Errorf("unexpected token[0]: pos=%d term=%q", output[0].Position, output[0].Term)
	}
	if output[1].Position != 1 || string(output[1].Term) != "procedure" {
		t.Errorf("unexpected token[1]: pos=%d term=%q", output[1].Position, output[1].Term)
	}
	if output[2].Position != 2 || string(output[2].Term) != "обновить" {
		t.Errorf("unexpected token[2]: pos=%d term=%q", output[2].Position, output[2].Term)
	}
	if output[3].Position != 3 || string(output[3].Term) != "если" {
		t.Errorf("unexpected token[3]: pos=%d term=%q", output[3].Position, output[3].Term)
	}
	if output[4].Position != 3 || string(output[4].Term) != "if" {
		t.Errorf("unexpected token[4]: pos=%d term=%q", output[4].Position, output[4].Term)
	}
}

func TestBuildSynonymMap_Bidirectional(t *testing.T) {
	m := buildSynonymMap()
	for k, v := range m {
		if k == v {
			t.Errorf("self-mapping: %q -> %q", k, v)
			continue
		}
		reverse, ok := m[v]
		if !ok {
			t.Errorf("broken: %q -> %q, but %q not in map", k, v, v)
		} else if reverse != k {
			t.Errorf("broken: %q -> %q -> %q (expected %q)", k, v, reverse, k)
		}
	}
}

func TestBuildBSLMapping(t *testing.T) {
	m := buildBSLMapping()

	// Default mapping must be disabled.
	if m.DefaultMapping.Enabled {
		t.Error("expected default mapping to be disabled")
	}

	// "module" document mapping must exist and be enabled.
	dm, ok := m.TypeMapping["module"]
	if !ok {
		t.Fatal("expected 'module' document mapping")
	}
	if !dm.Enabled {
		t.Error("expected 'module' mapping to be enabled")
	}

	// Verify field mappings exist.
	for _, field := range []string{"name", "category", "module", "content"} {
		prop, ok := dm.Properties[field]
		if !ok {
			t.Errorf("missing field mapping for %q", field)
			continue
		}
		if len(prop.Fields) == 0 {
			t.Errorf("field %q has no field mappings", field)
		}
	}

	// Verify "bsl" custom analyzer is registered.
	if _, ok := m.CustomAnalysis.Analyzers["bsl"]; !ok {
		t.Error("expected 'bsl' custom analyzer in mapping")
	}
}
