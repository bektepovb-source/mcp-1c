package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleBSLHelp_Found(t *testing.T) {
	input := BSLHelpInput{Query: "СтрНайти"}
	result, _, err := HandleBSLHelp(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "СтрНайти") {
		t.Fatalf("expected result to contain СтрНайти, got: %s", text)
	}
	if !strings.Contains(text, "StrFind") {
		t.Fatalf("expected result to contain StrFind, got: %s", text)
	}
}

func TestHandleBSLHelp_NotFound(t *testing.T) {
	input := BSLHelpInput{Query: "НесуществующаяФункция"}
	result, _, err := HandleBSLHelp(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for not-found case")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content for not-found case")
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "не найдена") {
		t.Fatalf("expected not-found message, got: %s", text)
	}
}

func TestHandleBSLHelp_EnglishQuery(t *testing.T) {
	input := BSLHelpInput{Query: "Format"}
	result, _, err := HandleBSLHelp(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "Формат") {
		t.Fatalf("expected result to contain Формат, got: %s", text)
	}
}

func TestBSLHelpTool(t *testing.T) {
	tool := BSLHelpTool()
	if tool.Name != "bsl_syntax_help" {
		t.Fatalf("expected tool name bsl_syntax_help, got %s", tool.Name)
	}
	if tool.Description == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestRegisterBSLHelp(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	// Should not panic.
	RegisterBSLHelp(s)
}
