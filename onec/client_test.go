package onec

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:8080/1c-mcp", "", "")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.BaseURL != "http://localhost:8080/1c-mcp" {
		t.Fatalf("expected base URL, got %s", c.BaseURL)
	}
	if c.User != "" {
		t.Fatalf("expected empty user, got %s", c.User)
	}
}

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/test" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "unexpected path", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"key":"value"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "")
	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["key"] != "value" {
		t.Fatalf("expected value, got %s", result["key"])
	}
}

func TestClientBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Error("expected basic auth header")
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}
		if user != "admin" || pass != "secret" {
			t.Errorf("unexpected credentials: %s:%s", user, pass)
			http.Error(w, "bad credentials", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "admin", "secret")
	var result map[string]any
	err := client.Get(context.Background(), "/auth", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientNoAuthWhenUserEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("expected no Authorization header when user is empty")
			http.Error(w, "unexpected auth header", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "")
	var result map[string]any
	err := client.Get(context.Background(), "/noauth", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientGetError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "", "")
	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
