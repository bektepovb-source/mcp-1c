package bsl

import "testing"

func TestBuiltinFunctionsCount(t *testing.T) {
	if len(BuiltinFunctions) < 120 {
		t.Fatalf("expected at least 120 builtin functions, got %d", len(BuiltinFunctions))
	}
}

func TestSearchByRussianName(t *testing.T) {
	results := Search("СтрНайти")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "СтрНайти" {
		t.Fatalf("expected СтрНайти, got %s", results[0].Name)
	}
}

func TestSearchByEnglishName(t *testing.T) {
	results := Search("StrFind")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	results := Search("strfind")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for case-insensitive search, got %d", len(results))
	}
}

func TestSearchNotFound(t *testing.T) {
	results := Search("НесуществующаяФункция")
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearchPartialMatch(t *testing.T) {
	results := Search("Формат")
	if len(results) < 1 {
		t.Fatal("expected at least 1 result for partial match")
	}
}

func TestSearchNewFunction_CurrentDate(t *testing.T) {
	results := Search("ТекущаяДата")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for ТекущаяДата, got %d", len(results))
	}
	if results[0].NameEn != "CurrentDate" {
		t.Fatalf("expected NameEn CurrentDate, got %s", results[0].NameEn)
	}
}

func TestSearchNewFunction_ValueIsFilled(t *testing.T) {
	results := Search("ValueIsFilled")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for ValueIsFilled, got %d", len(results))
	}
	if results[0].Name != "ЗначениеЗаполнено" {
		t.Fatalf("expected ЗначениеЗаполнено, got %s", results[0].Name)
	}
}

func TestSearchNewFunction_BeginTransaction(t *testing.T) {
	results := Search("НачатьТранзакцию")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for НачатьТранзакцию, got %d", len(results))
	}
	if results[0].NameEn != "BeginTransaction" {
		t.Fatalf("expected NameEn BeginTransaction, got %s", results[0].NameEn)
	}
}

func TestSearchStringFunctions(t *testing.T) {
	results := Search("Стр")
	if len(results) < 8 {
		t.Fatalf("expected at least 8 string functions matching 'Стр', got %d", len(results))
	}
}

func TestSearchNewFunction_BegOfDay(t *testing.T) {
	results := Search("НачалоДня")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for НачалоДня, got %d", len(results))
	}
	if results[0].NameEn != "BegOfDay" {
		t.Fatalf("expected NameEn BegOfDay, got %s", results[0].NameEn)
	}
}

func TestSearchNewFunction_StrStartsWith(t *testing.T) {
	results := Search("StrStartsWith")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for StrStartsWith, got %d", len(results))
	}
	if results[0].Name != "СтрНачинаетсяС" {
		t.Fatalf("expected СтрНачинаетсяС, got %s", results[0].Name)
	}
}

func TestSearchNewFunction_ErrorDescription(t *testing.T) {
	results := Search("ОписаниеОшибки")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for ОписаниеОшибки, got %d", len(results))
	}
	if results[0].NameEn != "ErrorDescription" {
		t.Fatalf("expected NameEn ErrorDescription, got %s", results[0].NameEn)
	}
}

func TestAllFunctionsHaveRequiredFields(t *testing.T) {
	for _, f := range BuiltinFunctions {
		if f.Name == "" {
			t.Error("found function with empty Name")
		}
		if f.NameEn == "" {
			t.Errorf("function %s has empty NameEn", f.Name)
		}
		if f.Description == "" {
			t.Errorf("function %s has empty Description", f.Name)
		}
		if f.Syntax == "" {
			t.Errorf("function %s has empty Syntax", f.Name)
		}
		if f.Example == "" {
			t.Errorf("function %s has empty Example", f.Name)
		}
	}
}
