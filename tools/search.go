package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/dump"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultSearchLimit = 50
	maxSearchLimit     = 500
)

// SearchCodeTool returns the MCP tool definition for search_code.
func SearchCodeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "search_code",
		Title: "Поиск по коду модулей",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		Description: "Глобальный поиск по тексту всех модулей конфигурации 1С (grep по коду). " +
			"Ищет подстроку и возвращает совпадения с контекстом. " +
			"Используй когда нужно найти где вызывается процедура, функция, переменная или обработчик. " +
			"Работает по локальной выгрузке конфигурации (DumpConfigToFiles).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {
					"type": "string",
					"description": "Строка для поиска в коде модулей"
				},
				"limit": {
					"type": "integer",
					"description": "Максимальное количество результатов (по умолчанию 50)"
				}
			},
			"required": ["query"]
		}`),
	}
}

// NewSearchCodeHandler returns a ToolHandler that searches BSL code in a local dump.
func NewSearchCodeHandler(searcher *dump.Searcher) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input queryLimitInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if input.Query == "" {
			return nil, fmt.Errorf("query is required")
		}
		input.Limit = clampLimit(input.Limit, defaultSearchLimit, maxSearchLimit)

		matches, total := searcher.Search(input.Query, input.Limit)
		return textResult(formatSearchResult(matches, total, input.Query)), nil
	}
}

func formatSearchResult(matches []dump.Match, total int, query string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## Результаты поиска \"%s\" (%d совпадений)\n\n", query, total)

	if len(matches) == 0 {
		b.WriteString("Ничего не найдено.\n")
		return b.String()
	}

	for _, m := range matches {
		fmt.Fprintf(&b, "### %s (строка %d)\n", m.Module, m.Line)
		b.WriteString("```bsl\n")
		b.WriteString(m.Context)
		b.WriteString("\n```\n\n")
	}

	if total > len(matches) {
		fmt.Fprintf(&b, "> Показано %d из %d совпадений. Уточните поиск или увеличьте limit.\n", len(matches), total)
	}

	return b.String()
}
