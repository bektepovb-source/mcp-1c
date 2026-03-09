package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestConfigurationInfoHandler(t *testing.T) {
	const mockResponse = `{
		"name": "БухгалтерияПредприятия",
		"version": "3.0.150.28",
		"vendor": "Фирма \"1С\"",
		"platform_version": "8.3.25.1394",
		"mode": "file"
	}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/configuration" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	client := onec.NewClient(mockServer.URL, "", "")
	handler := NewConfigurationInfoHandler(client)

	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name: "get_configuration_info",
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	for _, want := range []string{
		"БухгалтерияПредприятия",
		"3.0.150.28",
		"8.3.25.1394",
		"Файловый",
	} {
		if !strings.Contains(tc.Text, want) {
			t.Errorf("expected text to contain %q, got:\n%s", want, tc.Text)
		}
	}
}

func TestConfigurationInfoTool(t *testing.T) {
	tool := ConfigurationInfoTool()
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.Name != "get_configuration_info" {
		t.Errorf("expected tool name %q, got %q", "get_configuration_info", tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestFormatConfigurationInfo_ServerMode(t *testing.T) {
	info := &onec.ConfigurationInfo{
		Name:            "УправлениеТорговлей",
		Version:         "11.5.20.1",
		Vendor:          "1С",
		PlatformVersion: "8.5.1.100",
		Mode:            "server",
	}
	text := formatConfigurationInfo(info)
	if !strings.Contains(text, "Клиент-серверный") {
		t.Errorf("expected 'Клиент-серверный' for server mode, got:\n%s", text)
	}
}
