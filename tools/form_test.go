package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestFormStructureTool(t *testing.T) {
	tool := FormStructureTool()
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.Name != "get_form_structure" {
		t.Errorf("expected tool name %q, got %q", "get_form_structure", tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestFormatFormStructure(t *testing.T) {
	f := &onec.FormStructure{
		Name:  "ФормаДокумента",
		Title: "Реализация товаров и услуг",
		Elements: []onec.FormElement{
			{
				Name:     "Контрагент",
				Type:     "ПолеВвода",
				Title:    "Контрагент",
				DataPath: "Объект.Контрагент",
			},
			{
				Name:     "Сумма",
				Type:     "ПолеВвода",
				Title:    "Сумма документа",
				DataPath: "Объект.СуммаДокумента",
			},
			{
				Name:     "Товары",
				Type:     "ТаблицаФормы",
				Title:    "Товары",
				DataPath: "Объект.Товары",
			},
		},
		Commands: []onec.FormCommand{
			{Name: "Провести", Action: "Провести"},
			{Name: "ПечатьНакладной", Action: "ПечатьНакладной"},
		},
		Handlers: []onec.FormHandler{
			{Event: "ПриОткрытии", Handler: "ПриОткрытии"},
			{Event: "ПередЗаписью", Handler: "ПередЗаписью"},
		},
	}

	text := formatFormStructure(f)

	for _, want := range []string{
		"# Форма: ФормаДокумента",
		"Реализация товаров и услуг",
		"## Элементы формы",
		"| Контрагент | ПолеВвода | Контрагент | Объект.Контрагент |",
		"| Сумма | ПолеВвода | Сумма документа | Объект.СуммаДокумента |",
		"| Товары | ТаблицаФормы | Товары | Объект.Товары |",
		"## Команды формы",
		"**Провести**",
		"**ПечатьНакладной**",
		"## Обработчики событий",
		"**ПриОткрытии** → ПриОткрытии()",
		"**ПередЗаписью** → ПередЗаписью()",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("expected text to contain %q, got:\n%s", want, text)
		}
	}
}

func TestFormatFormStructure_Empty(t *testing.T) {
	f := &onec.FormStructure{
		Name: "ПустаяФорма",
	}

	text := formatFormStructure(f)

	if !strings.Contains(text, "# Форма: ПустаяФорма") {
		t.Errorf("expected form name in text, got:\n%s", text)
	}
	for _, section := range []string{
		"## Элементы формы",
		"## Команды формы",
		"## Обработчики событий",
	} {
		if strings.Contains(text, section) {
			t.Errorf("expected no %q section for empty form, got:\n%s", section, text)
		}
	}
}

func TestNewFormStructureHandler(t *testing.T) {
	mockForm := onec.FormStructure{
		Name:  "ФормаДокумента",
		Title: "Реализация товаров и услуг",
		Elements: []onec.FormElement{
			{
				Name:     "Контрагент",
				Type:     "ПолеВвода",
				Title:    "Контрагент",
				DataPath: "Объект.Контрагент",
			},
		},
		Commands: []onec.FormCommand{
			{Name: "Провести", Action: "Провести"},
		},
		Handlers: []onec.FormHandler{
			{Event: "ПриОткрытии", Handler: "ПриОткрытии"},
		},
	}
	mockResponse, _ := json.Marshal(mockForm)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/form/Document/РеализацияТоваровУслуг" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(mockResponse)
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewFormStructureHandler(client)

	args, _ := json.Marshal(map[string]any{
		"object_type": "Document",
		"object_name": "РеализацияТоваровУслуг",
	})
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "get_form_structure",
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
		"ФормаДокумента",
		"Реализация товаров и услуг",
		"Контрагент",
		"ПолеВвода",
		"Провести",
		"ПриОткрытии",
	} {
		if !strings.Contains(tc.Text, want) {
			t.Errorf("expected text to contain %q, got:\n%s", want, tc.Text)
		}
	}
}
