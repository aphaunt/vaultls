package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockPromoteServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/metadata/dev":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"DB_PASS", "API_KEY"}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/metadata/prod":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"API_KEY"}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/data/dev":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"DB_PASS": "secret", "API_KEY": "abc123"},
				},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/v1/secret/data/prod":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestPromoteSecrets_SamePath(t *testing.T) {
	client := &Client{Address: "http://localhost", Token: "tok"}
	_, err := PromoteSecrets(context.Background(), client, "secret/dev", "secret/dev", false)
	if err == nil {
		t.Fatal("expected error for same path")
	}
}

func TestPromoteSecrets_NoOverwrite(t *testing.T) {
	srv := newMockPromoteServer(t)
	defer srv.Close()
	client, _ := NewClient(srv.URL, "test-token")

	result, err := PromoteSecrets(context.Background(), client, "secret/dev", "secret/prod", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Copied) != 1 || result.Copied[0] != "DB_PASS" {
		t.Errorf("expected DB_PASS copied, got %v", result.Copied)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "API_KEY" {
		t.Errorf("expected API_KEY skipped, got %v", result.Skipped)
	}
}

func TestPromoteSecrets_WithOverwrite(t *testing.T) {
	srv := newMockPromoteServer(t)
	defer srv.Close()
	client, _ := NewClient(srv.URL, "test-token")

	result, err := PromoteSecrets(context.Background(), client, "secret/dev", "secret/prod", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Overwritten) != 1 || result.Overwritten[0] != "API_KEY" {
		t.Errorf("expected API_KEY overwritten, got %v", result.Overwritten)
	}
	if len(result.Copied) != 1 {
		t.Errorf("expected 1 copied key, got %v", result.Copied)
	}
}
