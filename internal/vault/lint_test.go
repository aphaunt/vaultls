package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockLintServer(data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"data": data,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestLintSecrets_EmptyPath(t *testing.T) {
	client, _ := NewClient("http://localhost", "token")
	_, err := LintSecrets(context.Background(), client, "", LintOptions{})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestLintSecrets_EmptyValues(t *testing.T) {
	server := newMockLintServer(map[string]interface{}{
		"db_password": "",
		"api_key":     "abc123",
	})
	defer server.Close()

	client, _ := NewClient(server.URL, "test-token")
	results, err := LintSecrets(context.Background(), client, "secret/data/app", LintOptions{
		DisallowEmptyValues: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 lint result, got %d", len(results))
	}
	if results[0].Key != "db_password" {
		t.Errorf("expected lint on db_password, got %s", results[0].Key)
	}
	if results[0].Severity != "error" {
		t.Errorf("expected severity error, got %s", results[0].Severity)
	}
}

func TestLintSecrets_UppercaseKeys(t *testing.T) {
	server := newMockLintServer(map[string]interface{}{
		"DB_HOST": "localhost",
		"api_key": "secret",
	})
	defer server.Close()

	client, _ := NewClient(server.URL, "test-token")
	results, err := LintSecrets(context.Background(), client, "secret/data/app", LintOptions{
		DisallowUppercaseKeys: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 lint result, got %d", len(results))
	}
	if results[0].Key != "DB_HOST" {
		t.Errorf("expected lint on DB_HOST, got %s", results[0].Key)
	}
}

func TestLintSecrets_RequirePrefix(t *testing.T) {
	server := newMockLintServer(map[string]interface{}{
		"app_name": "myapp",
		"version":  "1.0",
	})
	defer server.Close()

	client, _ := NewClient(server.URL, "test-token")
	results, err := LintSecrets(context.Background(), client, "secret/data/app", LintOptions{
		RequirePrefix: "app_",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for missing prefix, got %d", len(results))
	}
	if results[0].Key != "version" {
		t.Errorf("expected lint on 'version', got %s", results[0].Key)
	}
}

func TestLintSecrets_NoViolations(t *testing.T) {
	server := newMockLintServer(map[string]interface{}{
		"app_host": "localhost",
		"app_port": "8080",
	})
	defer server.Close()

	client, _ := NewClient(server.URL, "test-token")
	results, err := LintSecrets(context.Background(), client, "secret/data/app", LintOptions{
		DisallowEmptyValues:   true,
		DisallowUppercaseKeys: true,
		RequirePrefix:         "app_",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no lint results, got %d", len(results))
	}
}
