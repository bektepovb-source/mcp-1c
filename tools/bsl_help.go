package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/bsl"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BSLHelpInput is the input schema for the bsl_syntax_help tool.
type BSLHelpInput struct {
	Query string `json:"query" jsonschema:"Название функции на русском или английском, например СтрНайти или StrFind"`
}

// BSLHelpTool returns the Tool definition for bsl_syntax_help.
func BSLHelpTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "bsl_syntax_help",
		Title: "Справочник функций языка 1С",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		Description: "Справка по 180 встроенным функциям, методам типов и объектным паттернам языка 1С (BSL): строки, числа, даты, коллекции, файлы, HTTP, XML/JSON, Base64, двоичные данные, транзакции, система, методы ТаблицаЗначений/Массив/Структура/Соответствие/Запрос, справочники, документы, регистры сведений, системные перечисления. " +
			"Используй когда нужно узнать синтаксис, параметры или пример использования функции/метода платформы 1С:Предприятие. Параметр: query, название функции на русском или английском. " +
			"Вызывай перед написанием кода на BSL для уточнения сигнатуры и параметров функции.",
	}
}

// HandleBSLHelp is the handler for the bsl_syntax_help tool.
func HandleBSLHelp(_ context.Context, _ *mcp.CallToolRequest, input BSLHelpInput) (*mcp.CallToolResult, any, error) {
	results := bsl.Search(input.Query)

	if len(results) == 0 {
		return textResult(fmt.Sprintf("Функция %q не найдена в справочнике BSL.", input.Query)), nil, nil
	}

	var sb strings.Builder
	for i, f := range results {
		if i > 0 {
			sb.WriteString("\n---\n\n")
		}
		fmt.Fprintf(&sb, "## %s / %s\n\n", f.Name, f.NameEn)
		fmt.Fprintf(&sb, "**Описание:** %s\n\n", f.Description)
		fmt.Fprintf(&sb, "**Синтаксис:** `%s`\n\n", f.Syntax)
		fmt.Fprintf(&sb, "**Параметры:** %s\n\n", f.Parameters)
		fmt.Fprintf(&sb, "**Возвращает:** %s\n\n", f.ReturnType)
		fmt.Fprintf(&sb, "**Пример:**\n```bsl\n%s\n```\n", f.Example)
	}

	return textResult(sb.String()), nil, nil
}

// RegisterBSLHelp registers the bsl_syntax_help tool on the given server.
func RegisterBSLHelp(s *mcp.Server) {
	mcp.AddTool(s, BSLHelpTool(), HandleBSLHelp)
}
