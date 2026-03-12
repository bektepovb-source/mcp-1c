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
		Name:        "search_code",
		Title:       "Поиск по коду модулей",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		Description: "Полнотекстовый поиск по коду всех модулей конфигурации 1С. " +
			"Поддерживает три режима: smart (полнотекстовый с ранжированием BM25, " +
			"по умолчанию), regex (регулярные выражения), exact (точная подстрока). " +
			"Фильтрация по типу метаданных (category) и типу модуля (module). " +
			"BSL-синонимы: поиск по английским именам находит русские и наоборот " +
			"(StrFind -> СтрНайти, Procedure -> Процедура). " +
			"Работает по локальной выгрузке конфигурации (DumpConfigToFiles).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {
					"type": "string",
					"description": "Поисковый запрос. В режиме smart — слова для полнотекстового поиска. В режиме regex — регулярное выражение (Go regexp). В режиме exact — точная подстрока (регистронезависимо)."
				},
				"limit": {
					"type": "integer",
					"description": "Максимальное количество результатов (по умолчанию 50, максимум 500)"
				},
				"category": {
					"type": "string",
					"description": "Фильтр по типу метаданных: Документ, Справочник, ОбщийМодуль, Обработка, Отчет, РегистрСведений, РегистрНакопления и т.д. Значение чувствительно к регистру (например, 'Документ', не 'документ')."
				},
				"module": {
					"type": "string",
					"description": "Фильтр по типу модуля: МодульОбъекта, МодульМенеджера, МодульФормы, МодульНабораЗаписей, МодульКоманды, Модуль. Значение чувствительно к регистру (например, 'МодульОбъекта', не 'модульобъекта')."
				},
				"mode": {
					"type": "string",
					"enum": ["smart", "regex", "exact"],
					"description": "Режим поиска. smart — полнотекстовый с BM25-ранжированием и поддержкой BSL-синонимов (по умолчанию). regex — регулярное выражение. exact — точная подстрока."
				}
			},
			"required": ["query"]
		}`),
	}
}

// NewSearchCodeHandler returns a ToolHandler that searches BSL code in a local dump.
func NewSearchCodeHandler(index *dump.Index) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input searchCodeInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if input.Query == "" {
			return nil, fmt.Errorf("query is required")
		}

		limit := clampLimit(input.Limit, defaultSearchLimit, maxSearchLimit)
		var mode dump.SearchMode
		switch input.Mode {
		case "regex":
			mode = dump.SearchModeRegex
		case "exact":
			mode = dump.SearchModeExact
		case "smart", "":
			mode = dump.SearchModeSmart
		default:
			return nil, fmt.Errorf("unknown mode: %q (allowed: smart, regex, exact)", input.Mode)
		}

		matches, total, err := index.Search(dump.SearchParams{
			Query:    input.Query,
			Category: input.Category,
			Module:   input.Module,
			Mode:     mode,
			Limit:    limit,
		})
		if err != nil {
			return nil, fmt.Errorf("search: %w", err)
		}

		return textResult(formatSearchResult(matches, total, input.Query, mode)), nil
	}
}

func formatSearchResult(matches []dump.Match, total int, query string, mode dump.SearchMode) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## Результаты поиска \"%s\" (%d совпадений)\n\n", query, total)

	if len(matches) == 0 {
		b.WriteString("Ничего не найдено.\n")
		return b.String()
	}

	for _, m := range matches {
		if mode == dump.SearchModeSmart && m.Score > 0 {
			fmt.Fprintf(&b, "### %s (строка %d, score: %.3f)\n", m.Module, m.Line, m.Score)
		} else {
			fmt.Fprintf(&b, "### %s (строка %d)\n", m.Module, m.Line)
		}
		b.WriteString("```bsl\n")
		b.WriteString(m.Context)
		b.WriteString("\n```\n\n")
	}

	if total > len(matches) {
		fmt.Fprintf(&b, "> Показано %d из %d совпадений. Уточните поиск или увеличьте limit.\n", len(matches), total)
	}

	return b.String()
}
