package dump

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FormInfo holds parsed form structure from a dump XML file.
type FormInfo struct {
	Name     string
	Title    string
	Elements []FormElementInfo
	Commands []FormCommandInfo
	Handlers []FormHandlerInfo
}

// FormElementInfo represents a parsed form element.
type FormElementInfo struct {
	Name     string
	Type     string // XML element tag name (InputField, Table, etc.)
	Title    string
	DataPath string
}

// FormCommandInfo represents a parsed form command.
type FormCommandInfo struct {
	Name   string
	Action string
}

// FormHandlerInfo represents a parsed form event handler.
type FormHandlerInfo struct {
	Event   string
	Handler string
}

// objectTypeToDumpDir is defined in metadata_types.go and maps 1C object type
// names (as used in the tool input) to dump directory names.

// FindFormFiles locates all Form.xml files for the given object in the dump directory.
// It returns a map of form name to absolute file path.
func FindFormFiles(dumpDir, objectType, objectName string) (map[string]string, error) {
	dirName, ok := objectTypeToDumpDir[objectType]
	if !ok {
		return nil, fmt.Errorf("unknown object type %q for dump lookup", objectType)
	}

	if strings.Contains(objectName, "..") ||
		strings.Contains(objectName, "/") ||
		strings.Contains(objectName, "\\") {
		return nil, fmt.Errorf("invalid object name %q: contains path traversal characters", objectName)
	}

	formsDir := filepath.Join(dumpDir, dirName, objectName, "Forms")
	entries, err := os.ReadDir(formsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No forms directory — not an error.
		}
		return nil, fmt.Errorf("reading forms directory: %w", err)
	}

	result := make(map[string]string)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		formXML := filepath.Join(formsDir, entry.Name(), "Ext", "Form.xml")
		if _, statErr := os.Stat(formXML); statErr == nil {
			result[entry.Name()] = formXML
		}
	}

	return result, nil
}

// ParseFormXML parses a 1C form XML file and extracts elements, commands, and handlers.
func ParseFormXML(path string) (*FormInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading form XML: %w", err)
	}

	return parseFormXMLData(data)
}

// parseFormXMLData parses XML data from a 1C form dump file.
// The 1C form XML uses namespaces heavily. Go's xml.Decoder resolves
// prefixed names (e.g., v8:item) into their full namespace form, so we
// match on the Local part of the name only.
func parseFormXMLData(data []byte) (*FormInfo, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	form := &FormInfo{}

	// depth tracks XML nesting depth.
	depth := 0

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			local := t.Name.Local

			// Form > Title (depth 2).
			if depth == 2 && local == "Title" {
				form.Title = readLocalizedString(decoder, &depth)
			}

			// Form > Elements container (depth 2).
			if depth == 2 && local == "Elements" {
				form.Elements = parseElementsSection(decoder, &depth)
			}

			// Form > Commands container (depth 2).
			if depth == 2 && local == "Commands" {
				form.Commands = parseCommandsSection(decoder, &depth)
			}

			// Form > Handlers container (depth 2).
			if depth == 2 && local == "Handlers" {
				form.Handlers = parseHandlersSection(decoder, &depth)
			}

		case xml.EndElement:
			depth--
		}
	}

	return form, nil
}

// parseElementsSection reads all form elements from the <Elements> section,
// including nested elements inside groups and ChildItems.
func parseElementsSection(decoder *xml.Decoder, depth *int) []FormElementInfo {
	var elements []FormElementInfo
	sectionDepth := *depth

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			*depth++
			local := t.Name.Local

			if isFormElementTag(local) {
				elem, nested := parseFormElement(decoder, local, depth)
				elements = append(elements, elem)
				elements = append(elements, nested...)
			}

		case xml.EndElement:
			*depth--
			if *depth < sectionDepth {
				return elements
			}
		}
	}

	return elements
}

// parseFormElement reads a single form element and any nested child elements.
// It returns the element itself and a slice of any nested elements found in ChildItems.
func parseFormElement(decoder *xml.Decoder, elementType string, depth *int) (FormElementInfo, []FormElementInfo) {
	elem := FormElementInfo{Type: elementType}
	var nested []FormElementInfo
	elemDepth := *depth

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			*depth++
			local := t.Name.Local

			// Direct children of the element.
			if *depth == elemDepth+1 {
				switch local {
				case "Name":
					elem.Name = readCharData(decoder, depth)
				case "Title":
					elem.Title = readLocalizedString(decoder, depth)
				case "DataPath":
					elem.DataPath = readCharData(decoder, depth)
				case "ChildItems":
					nested = parseChildItems(decoder, depth)
				default:
					skipElement(decoder, depth)
				}
			}

		case xml.EndElement:
			*depth--
			if *depth < elemDepth {
				return elem, nested
			}
		}
	}

	return elem, nested
}

// parseChildItems reads nested form elements from a <ChildItems> section.
func parseChildItems(decoder *xml.Decoder, depth *int) []FormElementInfo {
	var elements []FormElementInfo
	childDepth := *depth

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			*depth++
			local := t.Name.Local

			if isFormElementTag(local) {
				elem, nested := parseFormElement(decoder, local, depth)
				elements = append(elements, elem)
				elements = append(elements, nested...)
			}

		case xml.EndElement:
			*depth--
			if *depth < childDepth {
				return elements
			}
		}
	}

	return elements
}

// parseCommandsSection reads all commands from the <Commands> section.
func parseCommandsSection(decoder *xml.Decoder, depth *int) []FormCommandInfo {
	var commands []FormCommandInfo
	sectionDepth := *depth

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			*depth++
			if t.Name.Local == "Command" {
				cmd := parseFormCommand(decoder, depth)
				commands = append(commands, cmd)
			}

		case xml.EndElement:
			*depth--
			if *depth < sectionDepth {
				return commands
			}
		}
	}

	return commands
}

// parseFormCommand reads a single form command.
func parseFormCommand(decoder *xml.Decoder, depth *int) FormCommandInfo {
	cmd := FormCommandInfo{}
	cmdDepth := *depth

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			*depth++
			if *depth == cmdDepth+1 {
				switch t.Name.Local {
				case "Name":
					cmd.Name = readCharData(decoder, depth)
				case "Action":
					cmd.Action = readCharData(decoder, depth)
				default:
					skipElement(decoder, depth)
				}
			}

		case xml.EndElement:
			*depth--
			if *depth < cmdDepth {
				return cmd
			}
		}
	}

	return cmd
}

// parseHandlersSection reads all event handlers from the <Handlers> section.
func parseHandlersSection(decoder *xml.Decoder, depth *int) []FormHandlerInfo {
	var handlers []FormHandlerInfo
	sectionDepth := *depth

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			*depth++
			if t.Name.Local == "Event" {
				h := parseFormHandler(decoder, depth)
				if h.Event != "" && h.Handler != "" {
					handlers = append(handlers, h)
				}
			}

		case xml.EndElement:
			*depth--
			if *depth < sectionDepth {
				return handlers
			}
		}
	}

	return handlers
}

// parseFormHandler reads a single event handler entry.
func parseFormHandler(decoder *xml.Decoder, depth *int) FormHandlerInfo {
	h := FormHandlerInfo{}
	handlerDepth := *depth

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			*depth++
			if *depth == handlerDepth+1 {
				switch t.Name.Local {
				case "Name":
					h.Event = readCharData(decoder, depth)
				case "Handler":
					h.Handler = readCharData(decoder, depth)
				default:
					skipElement(decoder, depth)
				}
			}

		case xml.EndElement:
			*depth--
			if *depth < handlerDepth {
				return h
			}
		}
	}

	return h
}

// readCharData reads the text content of the current element and consumes its end tag.
func readCharData(decoder *xml.Decoder, depth *int) string {
	var sb strings.Builder

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.CharData:
			sb.Write(t)
		case xml.StartElement:
			// Skip nested elements.
			*depth++
			skipElement(decoder, depth)
		case xml.EndElement:
			*depth--
			return strings.TrimSpace(sb.String())
		}
	}

	return strings.TrimSpace(sb.String())
}

// readLocalizedString reads a 1C localized string (v8:LocalStringType).
// It takes the first available value — typically the Russian text.
// In the XML, Go's decoder resolves "v8:item" to Local="item" with the namespace
// in Space, so we only match on Local.
func readLocalizedString(decoder *xml.Decoder, depth *int) string {
	var result string
	titleDepth := *depth

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" && result == "" {
				result = text
			}
		case xml.StartElement:
			*depth++
			local := t.Name.Local

			if local == "item" {
				val := readLocalizedItem(decoder, depth)
				if val != "" && result == "" {
					result = val
				}
			} else {
				skipElement(decoder, depth)
			}

		case xml.EndElement:
			*depth--
			if *depth < titleDepth {
				return result
			}
		}
	}

	return result
}

// readLocalizedItem reads a single <v8:item> with <v8:lang> and <v8:content> children.
func readLocalizedItem(decoder *xml.Decoder, depth *int) string {
	var content string
	itemDepth := *depth

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			*depth++
			if t.Name.Local == "content" {
				content = readCharData(decoder, depth)
			} else {
				skipElement(decoder, depth)
			}

		case xml.EndElement:
			*depth--
			if *depth < itemDepth {
				return content
			}
		}
	}

	return content
}

// skipElement consumes all tokens until the matching end element.
func skipElement(decoder *xml.Decoder, depth *int) {
	skipDepth := *depth

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch tok.(type) {
		case xml.StartElement:
			*depth++
		case xml.EndElement:
			*depth--
			if *depth < skipDepth {
				return
			}
		}
	}
}

// formElementTags lists XML tag names that represent form elements.
var formElementTags = map[string]bool{
	"InputField":              true,
	"LabelField":              true,
	"CheckBoxField":           true,
	"RadioButtonField":        true,
	"NumberField":             true,
	"TextDocumentField":       true,
	"SpreadsheetDocumentField": true,
	"PictureField":            true,
	"Table":                   true,
	"FormattedDocumentField":  true,
	"PlannerField":            true,
	"DendrogramField":         true,
	"ChartField":              true,
	"GanttChartField":         true,
	"PeriodField":             true,
	"ProgressBarField":        true,
	"TrackBarField":           true,
	"CalendarField":           true,
	"HTMLDocumentField":       true,
	"Button":                  true,
	"UsualGroup":              true,
	"Pages":                   true,
	"Page":                    true,
	"CommandBar":              true,
	"Popup":                   true,
	"ContextMenu":             true,
	"SearchControlAddition":   true,
	"ViewStatusAddition":      true,
	"SearchStringAddition":    true,
	"ColumnGroup":             true,
	"LabelDecoration":         true,
	"PictureDecoration":       true,
	"Hyperlink":               true,
	"Addition":                true,
}

func isFormElementTag(tag string) bool {
	return formElementTags[tag]
}

// elementTypeDisplayName maps XML element types to Russian display names.
var elementTypeDisplayName = map[string]string{
	"InputField":              "ПолеВвода",
	"LabelField":              "ПолеНадписи",
	"CheckBoxField":           "ФлажокПоле",
	"RadioButtonField":        "ПолеПереключателя",
	"NumberField":             "ПолеЧисла",
	"TextDocumentField":       "ПолеТекстовогоДокумента",
	"SpreadsheetDocumentField": "ПолеТабличногоДокумента",
	"PictureField":            "ПолеКартинки",
	"Table":                   "ТаблицаФормы",
	"Button":                  "Кнопка",
	"UsualGroup":              "ОбычнаяГруппа",
	"Pages":                   "Страницы",
	"Page":                    "Страница",
	"CommandBar":              "КоманднаяПанель",
	"LabelDecoration":         "ДекорацияНадпись",
	"PictureDecoration":       "ДекорацияКартинка",
	"Hyperlink":               "Гиперссылка",
}

// DisplayType returns a Russian name for the element type, or the raw tag if unknown.
func DisplayType(elementType string) string {
	if name, ok := elementTypeDisplayName[elementType]; ok {
		return name
	}
	return elementType
}
