package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

// textResult wraps a text string into an MCP tool result.
func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

// objectInput is the common input for tools that operate on a specific metadata object.
type objectInput struct {
	ObjectType string `json:"object_type"`
	ObjectName string `json:"object_name"`
}

// clampLimit normalises a user-supplied limit to [defaultVal, maxVal].
func clampLimit(value, defaultVal, maxVal int) int {
	if value <= 0 {
		return defaultVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}
