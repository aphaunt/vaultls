package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultls/internal/vault"
)

func newMockSearchServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Query().Get("list") == "true":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"db", "api"},
				},
			})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/db"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]string{
						"password": "supersecret",
						"host":     "db.internal",
					},
				},
			})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/api"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]string{
						"token": "abc123",
						"url":   "https://api.example.com",
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestSearchSecrets_MatchByKey(t *testing.T) {
	srv := newMockSearchServer(t)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	results, err := vault.SearchSecrets(context.Background(), client, "secret/", "password")
	if err != nil {
		t.Fatalf("SearchSecrets: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if _, ok := results[0].Matches["password"]; !ok {
		t.Error("expected 'password' key in matches")
	}
}

func TestSearchSecrets_NoMatch(t *testing.T) {
	srv := newMockSearchServer(t)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	results, err := vault.SearchSecrets(context.Background(), client, "secret/", "nonexistent")
	if err != nil {
		t.Fatalf("SearchSecrets: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
