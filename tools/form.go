package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FormStructureTool returns the MCP tool definition for get_form_structure.
func FormStructureTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "get_form_structure",
		Title: "Структура формы объекта",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		Description: "Получить структуру управляемой формы объекта 1С: элементы интерфейса, команды, кнопки и обработчики событий. " +
			"Используй когда нужно понять как выглядит форма документа, справочника или обработки.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"object_type": {
					"type": "string",
					"description": "Тип объекта: Document, Catalog, DataProcessor, Report и т.д."
				},
				"object_name": {
					"type": "string",
					"description": "Имя объекта метаданных"
				}
			},
			"required": ["object_type", "object_name"]
		}`),
	}
}

// NewFormStructureHandler returns a ToolHandler that fetches form structure from 1C.
func NewFormStructureHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input objectInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if input.ObjectType == "" || input.ObjectName == "" {
			return nil, fmt.Errorf("object_type and object_name are required")
		}

		endpoint := fmt.Sprintf("/form/%s/%s", input.ObjectType, input.ObjectName)
		var form onec.FormStructure
		if err := client.Get(ctx, endpoint, &form); err != nil {
			return nil, fmt.Errorf("fetching form structure from 1C: %w", err)
		}

		return textResult(formatFormStructure(&form)), nil
	}
}

func formatFormStructure(f *onec.FormStructure) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Форма: %s\n", f.Name)
	if f.Title != "" {
		fmt.Fprintf(&b, "**Заголовок:** %s\n", f.Title)
	}
	b.WriteByte('\n')

	if len(f.Elements) > 0 {
		b.WriteString("## Элементы формы\n\n")
		b.WriteString("| Имя | Тип | Заголовок | Путь к данным |\n")
		b.WriteString("|-----|-----|-----------|---------------|\n")
		for _, e := range f.Elements {
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", e.Name, e.Type, e.Title, e.DataPath)
		}
		b.WriteByte('\n')
	}

	if len(f.Commands) > 0 {
		b.WriteString("## Команды формы\n\n")
		for _, c := range f.Commands {
			fmt.Fprintf(&b, "- **%s** → %s\n", c.Name, c.Action)
		}
		b.WriteByte('\n')
	}

	if len(f.Handlers) > 0 {
		b.WriteString("## Обработчики событий\n\n")
		for _, h := range f.Handlers {
			fmt.Fprintf(&b, "- **%s** → %s()\n", h.Event, h.Handler)
		}
		b.WriteByte('\n')
	}

	return b.String()
}
