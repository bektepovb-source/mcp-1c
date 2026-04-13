package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/feenlace/mcp-1c/onec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DossierFilesTool returns the MCP tool definition for getting universal dossier files.
func DossierFilesTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "get_dossier_files",
		Description: "Get universal dossier files for any object. " +
			"The tool uses universal logic to find all 'ДокументыКредитногоДосье' related to the passed document, " +
			"and then collects all files from 'ФайлыКредитногоДосье'.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"object_type": {
					"type": "string",
					"description": "Тип объекта метаданных, например Документ.ДоговорЗалога"
				},
				"guid": {
					"type": "string",
					"description": "Уникальный идентификатор (GUID) объекта"
				}
			},
			"required": ["object_type", "guid"]
		}`),
	}
}

// dossierFilesInput is the input for the dossier files tool.
type dossierFilesInput struct {
	ObjectType string `json:"object_type"`
	GUID       string `json:"guid"`
}

// dossierFile represents a file returned from 1C.
type dossierFile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Size      int    `json:"size"`
	Data      string `json:"data"` // Base64 string
}

// NewDossierFilesHandler returns a ToolHandler that fetches dossier files from 1C.
func NewDossierFilesHandler(client *onec.Client) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input dossierFilesInput
		if req.Params.Arguments != nil {
			if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}
		}

		if input.ObjectType == "" || input.GUID == "" {
			return nil, fmt.Errorf("object_type and guid are required")
		}

		path := fmt.Sprintf("/dossier/%s/%s", input.ObjectType, input.GUID)

		var files []dossierFile
		if err := client.Get(ctx, path, &files); err != nil {
			return nil, fmt.Errorf("fetching dossier files from 1C: %w", err)
		}

		if len(files) == 0 {
			return textResult("No files found in the dossier for this object."), nil
		}

		var contents []mcp.Content

		for _, f := range files {
			ext := strings.ToLower(strings.TrimPrefix(f.Extension, "."))

			// Process data
			var fileData []byte
			if f.Data != "" {
				decoded, err := base64.StdEncoding.DecodeString(f.Data)
				if err == nil {
					fileData = decoded
				}
			}

			if len(fileData) > 0 {
				switch ext {
				case "png", "jpg", "jpeg", "gif", "bmp":
					mimeType := "image/" + ext
					if ext == "jpg" {
						mimeType = "image/jpeg"
					}
					contents = append(contents, &mcp.ImageContent{
						Data:     fileData,
						MIMEType: mimeType,
					})
					contents = append(contents, &mcp.TextContent{
						Text: fmt.Sprintf("Image file: %s", f.Name),
					})
				case "txt", "json", "xml", "csv", "md":
					contents = append(contents, &mcp.TextContent{
						Text: fmt.Sprintf("--- File: %s ---\n%s\n---", f.Name, string(fileData)),
					})
				default:
					contents = append(contents, &mcp.TextContent{
						Text: fmt.Sprintf("File: %s (Type: %s, Size: %d bytes) - binary content not displayed", f.Name, f.Extension, f.Size),
					})
				}
			} else {
				contents = append(contents, &mcp.TextContent{
					Text: fmt.Sprintf("File: %s (Type: %s, Size: %d bytes) - no content downloaded", f.Name, f.Extension, f.Size),
				})
			}
		}

		return &mcp.CallToolResult{
			Content: contents,
		}, nil
	}
}
