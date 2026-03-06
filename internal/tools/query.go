package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/feenlace/mcp-1c/internal/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultQueryLimit = 100
	maxQueryLimit     = 1000
)

type queryInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// QueryTool returns the MCP tool definition for execute_query.
func QueryTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "execute_query",
		Description: "Выполнить запрос на языке запросов 1С (только SELECT/ВЫБРАТЬ). Используй когда нужно получить данные из базы 1С: список контрагентов, остатки на складе, обороты по регистру и т.д.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {
					"type": "string",
					"description": "Текст запроса на языке запросов 1С. Только ВЫБРАТЬ/SELECT."
				},
				"limit": {
					"type": "integer",
					"description": "Максимальное количество строк результата (по умолчанию 100, максимум 1000)"
				}
			},
			"required": ["query"]
		}`),
	}
}

// NewQueryHandler returns a ToolHandler that executes a read-only query in 1C.
func NewQueryHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input queryInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if input.Query == "" {
			return nil, fmt.Errorf("query is required")
		}

		// Client-side read-only check (hint for LLM callers).
		// Server-side 1C extension enforces read-only execution.
		prefix := strings.TrimSpace(input.Query)
		if len(prefix) > 30 {
			prefix = prefix[:30]
		}
		upper := strings.ToUpper(prefix)
		if !strings.HasPrefix(upper, "ВЫБРАТЬ") && !strings.HasPrefix(upper, "SELECT") {
			return nil, fmt.Errorf("только SELECT/ВЫБРАТЬ запросы разрешены")
		}

		input.Limit = clampLimit(input.Limit, defaultQueryLimit, maxQueryLimit)

		body := onec.QueryRequest{
			Query: input.Query,
			Limit: input.Limit,
		}
		var result onec.QueryResult
		if err := client.Post(ctx, "/query", body, &result); err != nil {
			return nil, fmt.Errorf("executing query in 1C: %w", err)
		}

		return textResult(formatQueryResult(&result)), nil
	}
}

// formatQueryResult formats the query result as a markdown table.
func formatQueryResult(r *onec.QueryResult) string {
	var b strings.Builder
	b.Grow(len(r.Columns) * len(r.Rows) * 50)

	fmt.Fprintf(&b, "## Результат запроса (%d записей)\n\n", r.Total)

	if len(r.Columns) == 0 || len(r.Rows) == 0 {
		b.WriteString("Нет данных.\n")
		return b.String()
	}

	// Header
	b.WriteString("| ")
	b.WriteString(strings.Join(r.Columns, " | "))
	b.WriteString(" |\n")

	// Separator
	b.WriteString("|")
	for range r.Columns {
		b.WriteString("---|")
	}
	b.WriteByte('\n')

	// Rows
	for _, row := range r.Rows {
		b.WriteString("| ")
		for i, cell := range row {
			if i > 0 {
				b.WriteString(" | ")
			}
			b.WriteString(formatCell(cell))
		}
		b.WriteString(" |\n")
	}

	if r.Truncated {
		b.WriteString("\n> Результат усечён. Показаны первые записи. Используйте параметр `limit` для увеличения.\n")
	}

	return b.String()
}

// formatCell converts a JSON-deserialized cell value to string without reflection.
func formatCell(v any) string {
	switch c := v.(type) {
	case string:
		return c
	case float64:
		if c == float64(int64(c)) {
			return strconv.FormatInt(int64(c), 10)
		}
		return strconv.FormatFloat(c, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(c)
	case nil:
		return ""
	default:
		return fmt.Sprint(v)
	}
}
