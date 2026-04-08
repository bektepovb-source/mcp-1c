package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/feenlace/mcp-1c/dump"
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
				},
				"form_name": {
					"type": "string",
					"description": "Имя формы (если не указано — возвращается первая найденная форма)"
				}
			},
			"required": ["object_type", "object_name"]
		}`),
	}
}

// formInput extends objectInput with an optional form name.
type formInput struct {
	objectInput
	FormName string `json:"form_name"`
}

// NewFormStructureHandler returns a ToolHandler that fetches form structure.
// If dumpDir is provided, it will be used to enrich the result with data
// parsed from Form.xml files (elements, commands, handlers) which are not
// available through the 1C HTTP endpoint in Enterprise mode.
func NewFormStructureHandler(client *onec.Client, dumpDir string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input formInput
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("parsing input: %w", err)
		}
		if input.ObjectType == "" || input.ObjectName == "" {
			return nil, fmt.Errorf("object_type and object_name are required")
		}

		// Try 1C HTTP endpoint first.
		var form onec.FormStructure
		var httpErr error
		endpoint := fmt.Sprintf("/form/%s/%s", input.ObjectType, input.ObjectName)
		httpErr = client.Get(ctx, endpoint, &form)

		// If dump directory is configured, try to enrich from Form.xml.
		if dumpDir != "" {
			dumpForm, dumpErr := formFromDump(dumpDir, input.ObjectType, input.ObjectName, input.FormName)
			if dumpErr == nil && dumpForm != nil {
				// Use dump data if HTTP failed or returned empty elements/commands/handlers.
				if httpErr != nil {
					form = *dumpForm
				} else {
					enrichFormFromDump(&form, dumpForm)
				}
			} else if httpErr != nil {
				// Both sources failed.
				return nil, fmt.Errorf("fetching form structure from 1C: %w (dump fallback: %v)", httpErr, dumpErr)
			}
		} else if httpErr != nil {
			return nil, fmt.Errorf("fetching form structure from 1C: %w", httpErr)
		}

		return textResult(formatFormStructure(&form)), nil
	}
}

// formFromDump loads form structure from a DumpConfigToFiles XML file.
func formFromDump(dumpDir, objectType, objectName, formName string) (*onec.FormStructure, error) {
	formFiles, err := dump.FindFormFiles(dumpDir, objectType, objectName)
	if err != nil {
		return nil, fmt.Errorf("finding form files: %w", err)
	}
	if len(formFiles) == 0 {
		return nil, fmt.Errorf("no forms found in dump for %s.%s", objectType, objectName)
	}

	// Select the requested form or pick the first one.
	var selectedPath string
	var selectedName string
	if formName != "" {
		path, ok := formFiles[formName]
		if !ok {
			return nil, fmt.Errorf("form %q not found in dump (available: %s)", formName, joinMapKeys(formFiles))
		}
		selectedPath = path
		selectedName = formName
	} else {
		// Pick the first form alphabetically for deterministic results.
		keys := make([]string, 0, len(formFiles))
		for name := range formFiles {
			keys = append(keys, name)
		}
		slices.Sort(keys)
		selectedName = keys[0]
		selectedPath = formFiles[selectedName]
	}

	parsed, err := dump.ParseFormXML(selectedPath)
	if err != nil {
		return nil, fmt.Errorf("parsing form XML %q: %w", selectedPath, err)
	}

	return convertDumpForm(selectedName, parsed), nil
}

// convertDumpForm converts dump.FormInfo to onec.FormStructure.
func convertDumpForm(formName string, info *dump.FormInfo) *onec.FormStructure {
	form := &onec.FormStructure{
		Name:     formName,
		Title:    info.Title,
		Elements: make([]onec.FormElement, 0, len(info.Elements)),
		Commands: make([]onec.FormCommand, 0, len(info.Commands)),
		Handlers: make([]onec.FormHandler, 0, len(info.Handlers)),
	}

	for _, e := range info.Elements {
		form.Elements = append(form.Elements, onec.FormElement{
			Name:     e.Name,
			Type:     dump.DisplayType(e.Type),
			Title:    e.Title,
			DataPath: e.DataPath,
		})
	}

	for _, c := range info.Commands {
		form.Commands = append(form.Commands, onec.FormCommand{
			Name:   c.Name,
			Action: c.Action,
		})
	}

	for _, h := range info.Handlers {
		form.Handlers = append(form.Handlers, onec.FormHandler{
			Event:   h.Event,
			Handler: h.Handler,
		})
	}

	return form
}

// enrichFormFromDump fills in empty sections of the HTTP response with dump data.
func enrichFormFromDump(form *onec.FormStructure, dumpForm *onec.FormStructure) {
	if len(form.Elements) == 0 && len(dumpForm.Elements) > 0 {
		form.Elements = dumpForm.Elements
	}
	if len(form.Commands) == 0 && len(dumpForm.Commands) > 0 {
		form.Commands = dumpForm.Commands
	}
	if len(form.Handlers) == 0 && len(dumpForm.Handlers) > 0 {
		form.Handlers = dumpForm.Handlers
	}
	if form.Title == "" && dumpForm.Title != "" {
		form.Title = dumpForm.Title
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
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n",
				escapePipe(e.Name), escapePipe(e.Type), escapePipe(e.Title), escapePipe(e.DataPath))
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

// escapePipe escapes pipe characters so they do not break markdown tables.
func escapePipe(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

// joinMapKeys returns a comma-separated list of map keys.
func joinMapKeys(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return strings.Join(keys, ", ")
}
