package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/feenlace/mcp-1c/dump"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSearchCodeTool(t *testing.T) {
	tool := SearchCodeTool()
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.Name != "search_code" {
		t.Errorf("expected tool name %q, got %q", "search_code", tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestFormatSearchResult(t *testing.T) {
	matches := []dump.Match{
		{
			Module:  "Справочник.Контрагенты.МодульОбъекта",
			Line:    42,
			Context: "Процедура ПередЗаписью(Отказ)\n    // проверка заполнения\nКонецПроцедуры",
		},
		{
			Module:  "Документ.РеализацияТоваров.МодульОбъекта",
			Line:    15,
			Context: "Функция ПолучитьКонтрагента()\n    Возврат Контрагент;\nКонецФункции",
		},
	}

	text := formatSearchResult(matches, 2, "Контрагент")

	for _, want := range []string{
		"Результаты поиска",
		"Контрагент",
		"2 совпадений",
		"Справочник.Контрагенты.МодульОбъекта",
		"строка 42",
		"```bsl",
		"ПередЗаписью",
		"Документ.РеализацияТоваров.МодульОбъекта",
		"строка 15",
		"ПолучитьКонтрагента",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("expected text to contain %q, got:\n%s", want, text)
		}
	}
}

func TestFormatSearchResult_Empty(t *testing.T) {
	text := formatSearchResult(nil, 0, "НесуществующаяФункция")

	if !strings.Contains(text, "Ничего не найдено") {
		t.Errorf("expected 'Ничего не найдено' in text, got:\n%s", text)
	}
	if !strings.Contains(text, "0 совпадений") {
		t.Errorf("expected '0 совпадений' in text, got:\n%s", text)
	}
}

func TestFormatSearchResult_Truncated(t *testing.T) {
	matches := []dump.Match{
		{
			Module:  "Модуль.Тест",
			Line:    1,
			Context: "Тест",
		},
	}

	text := formatSearchResult(matches, 150, "Тест")

	if !strings.Contains(text, "Показано 1 из 150 совпадений") {
		t.Errorf("expected truncation message, got:\n%s", text)
	}
	if !strings.Contains(text, "увеличьте limit") {
		t.Errorf("expected limit hint in text, got:\n%s", text)
	}
}

func TestNewSearchCodeHandler(t *testing.T) {
	dir := t.TempDir()
	mkBSL(t, dir, "Catalogs/Номенклатура/Ext/ObjectModule.bsl",
		"Строка1\nСтрока2\nПроцедура ОбновитьЦены()\n    // обновление цен\nКонецПроцедуры\n")

	searcher, err := dump.NewSearcher(dir)
	if err != nil {
		t.Fatalf("NewSearcher: %v", err)
	}

	handler := NewSearchCodeHandler(searcher)

	args, _ := json.Marshal(map[string]any{
		"query": "ОбновитьЦены",
	})
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "search_code",
			Arguments: args,
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}

	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if tc.Text == "" {
		t.Fatal("expected non-empty text")
	}

	for _, want := range []string{
		"Справочник.Номенклатура.МодульОбъекта",
		"строка 3",
		"ОбновитьЦены",
		"1 совпадений",
	} {
		if !strings.Contains(tc.Text, want) {
			t.Errorf("expected text to contain %q, got:\n%s", want, tc.Text)
		}
	}
}

func mkBSL(t *testing.T, base, relPath, content string) {
	t.Helper()
	full := filepath.Join(base, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
