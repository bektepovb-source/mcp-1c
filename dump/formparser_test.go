package dump

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFormXML_Basic(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<Form xmlns="http://v8.1c.ru/8.3/xcf/logform"
      xmlns:v8="http://v8.1c.ru/8.1/data/core">
  <Title>
    <v8:item>
      <v8:lang>ru</v8:lang>
      <v8:content>Реализация товаров и услуг</v8:content>
    </v8:item>
  </Title>
  <Elements>
    <InputField>
      <Name>Контрагент</Name>
      <Title>
        <v8:item>
          <v8:lang>ru</v8:lang>
          <v8:content>Контрагент</v8:content>
        </v8:item>
      </Title>
      <DataPath>Объект.Контрагент</DataPath>
    </InputField>
    <Table>
      <Name>Товары</Name>
      <DataPath>Объект.Товары</DataPath>
    </Table>
    <Button>
      <Name>КнопкаПровести</Name>
      <Title>
        <v8:item>
          <v8:lang>ru</v8:lang>
          <v8:content>Провести</v8:content>
        </v8:item>
      </Title>
    </Button>
  </Elements>
  <Commands>
    <Command>
      <Name>Провести</Name>
      <Action>Провести</Action>
    </Command>
    <Command>
      <Name>ПечатьНакладной</Name>
      <Action>ПечатьНакладной</Action>
    </Command>
  </Commands>
  <Handlers>
    <Event>
      <Name>ПриОткрытии</Name>
      <Handler>ПриОткрытии</Handler>
    </Event>
    <Event>
      <Name>ПередЗаписью</Name>
      <Handler>ПередЗаписью</Handler>
    </Event>
  </Handlers>
</Form>`

	form, err := parseFormXMLData([]byte(xml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Title.
	if form.Title != "Реализация товаров и услуг" {
		t.Errorf("expected title %q, got %q", "Реализация товаров и услуг", form.Title)
	}

	// Elements.
	if len(form.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(form.Elements))
	}

	assertElement(t, form.Elements[0], "Контрагент", "InputField", "Контрагент", "Объект.Контрагент")
	assertElement(t, form.Elements[1], "Товары", "Table", "", "Объект.Товары")
	assertElement(t, form.Elements[2], "КнопкаПровести", "Button", "Провести", "")

	// Commands.
	if len(form.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(form.Commands))
	}
	if form.Commands[0].Name != "Провести" || form.Commands[0].Action != "Провести" {
		t.Errorf("unexpected command[0]: %+v", form.Commands[0])
	}
	if form.Commands[1].Name != "ПечатьНакладной" || form.Commands[1].Action != "ПечатьНакладной" {
		t.Errorf("unexpected command[1]: %+v", form.Commands[1])
	}

	// Handlers.
	if len(form.Handlers) != 2 {
		t.Fatalf("expected 2 handlers, got %d", len(form.Handlers))
	}
	if form.Handlers[0].Event != "ПриОткрытии" || form.Handlers[0].Handler != "ПриОткрытии" {
		t.Errorf("unexpected handler[0]: %+v", form.Handlers[0])
	}
	if form.Handlers[1].Event != "ПередЗаписью" || form.Handlers[1].Handler != "ПередЗаписью" {
		t.Errorf("unexpected handler[1]: %+v", form.Handlers[1])
	}
}

func TestParseFormXML_NestedElements(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<Form xmlns="http://v8.1c.ru/8.3/xcf/logform">
  <Elements>
    <UsualGroup>
      <Name>ГруппаОсновная</Name>
      <ChildItems>
        <InputField>
          <Name>Контрагент</Name>
          <DataPath>Объект.Контрагент</DataPath>
        </InputField>
      </ChildItems>
    </UsualGroup>
  </Elements>
</Form>`

	form, err := parseFormXMLData([]byte(xml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find both the group and the nested element.
	if len(form.Elements) < 2 {
		t.Fatalf("expected at least 2 elements (group + nested), got %d", len(form.Elements))
	}

	// Check that the group was parsed.
	foundGroup := false
	foundField := false
	for _, e := range form.Elements {
		if e.Name == "ГруппаОсновная" && e.Type == "UsualGroup" {
			foundGroup = true
		}
		if e.Name == "Контрагент" && e.Type == "InputField" {
			foundField = true
		}
	}
	if !foundGroup {
		t.Error("expected to find UsualGroup element")
	}
	if !foundField {
		t.Error("expected to find nested InputField element")
	}
}

func TestParseFormXML_Empty(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<Form xmlns="http://v8.1c.ru/8.3/xcf/logform">
</Form>`

	form, err := parseFormXMLData([]byte(xml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(form.Elements) != 0 {
		t.Errorf("expected 0 elements, got %d", len(form.Elements))
	}
	if len(form.Commands) != 0 {
		t.Errorf("expected 0 commands, got %d", len(form.Commands))
	}
	if len(form.Handlers) != 0 {
		t.Errorf("expected 0 handlers, got %d", len(form.Handlers))
	}
}

func TestParseFormXML_File(t *testing.T) {
	dir := t.TempDir()
	formDir := filepath.Join(dir, "Documents", "ТестДок", "Forms", "ФормаДокумента", "Ext")
	if err := os.MkdirAll(formDir, 0o755); err != nil {
		t.Fatal(err)
	}

	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<Form xmlns="http://v8.1c.ru/8.3/xcf/logform">
  <Elements>
    <InputField>
      <Name>Поле1</Name>
      <DataPath>Объект.Реквизит1</DataPath>
    </InputField>
  </Elements>
  <Commands>
    <Command>
      <Name>Команда1</Name>
      <Action>Действие1</Action>
    </Command>
  </Commands>
</Form>`

	xmlPath := filepath.Join(formDir, "Form.xml")
	if err := os.WriteFile(xmlPath, []byte(xmlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	form, err := ParseFormXML(xmlPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(form.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(form.Elements))
	}
	if form.Elements[0].Name != "Поле1" {
		t.Errorf("expected element name %q, got %q", "Поле1", form.Elements[0].Name)
	}
}

func TestFindFormFiles(t *testing.T) {
	dir := t.TempDir()

	// Create two form directories.
	for _, formName := range []string{"ФормаДокумента", "ФормаСписка"} {
		formDir := filepath.Join(dir, "Documents", "ТестДок", "Forms", formName, "Ext")
		if err := os.MkdirAll(formDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(formDir, "Form.xml"), []byte("<Form/>"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	forms, err := FindFormFiles(dir, "Document", "ТестДок")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(forms) != 2 {
		t.Fatalf("expected 2 forms, got %d", len(forms))
	}

	if _, ok := forms["ФормаДокумента"]; !ok {
		t.Error("expected ФормаДокумента in results")
	}
	if _, ok := forms["ФормаСписка"]; !ok {
		t.Error("expected ФормаСписка in results")
	}
}

func TestFindFormFiles_NoFormsDir(t *testing.T) {
	dir := t.TempDir()

	forms, err := FindFormFiles(dir, "Document", "НесуществующийДок")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if forms != nil {
		t.Errorf("expected nil for missing forms directory, got %v", forms)
	}
}

func TestFindFormFiles_UnknownType(t *testing.T) {
	dir := t.TempDir()

	_, err := FindFormFiles(dir, "UnknownType", "Test")
	if err == nil {
		t.Fatal("expected error for unknown object type")
	}
	if !strings.Contains(err.Error(), "unknown object type") {
		t.Errorf("expected 'unknown object type' in error, got: %v", err)
	}
}

func TestFindFormFiles_PathTraversal(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name       string
		objectName string
	}{
		{"dot-dot", "../../etc"},
		{"dot-dot-only", ".."},
		{"forward-slash", "foo/bar"},
		{"backslash", "foo\\bar"},
		{"dot-dot-backslash", "..\\secret"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := FindFormFiles(dir, "Document", tc.objectName)
			if err == nil {
				t.Fatal("expected error for path traversal in object name")
			}
			if !strings.Contains(err.Error(), "path traversal") {
				t.Errorf("expected 'path traversal' in error, got: %v", err)
			}
		})
	}
}

func TestDisplayType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"InputField", "ПолеВвода"},
		{"Table", "ТаблицаФормы"},
		{"Button", "Кнопка"},
		{"UsualGroup", "ОбычнаяГруппа"},
		{"UnknownElement", "UnknownElement"},
	}

	for _, tt := range tests {
		got := DisplayType(tt.input)
		if got != tt.want {
			t.Errorf("DisplayType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func assertElement(t *testing.T, e FormElementInfo, name, typ, title, dataPath string) {
	t.Helper()
	if e.Name != name {
		t.Errorf("expected element name %q, got %q", name, e.Name)
	}
	if e.Type != typ {
		t.Errorf("expected element type %q, got %q", typ, e.Type)
	}
	if title != "" && e.Title != title {
		t.Errorf("expected element title %q, got %q", title, e.Title)
	}
	if dataPath != "" && e.DataPath != dataPath {
		t.Errorf("expected element dataPath %q, got %q", dataPath, e.DataPath)
	}
}
