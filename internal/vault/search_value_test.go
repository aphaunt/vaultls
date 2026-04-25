package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockSearchValueServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/app":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"db", "api"}},
			})
		case "/v1/secret/data/app/db":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]string{"password": "supersecret", "user": "admin"}},
			})
		case "/v1/secret/data/app/api":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]string{"token": "abc123", "endpoint": "https://api.example.com"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestSearchSecretsByValue_Match(t *testing.T) {
	srv := newMockSearchValueServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	results, err := SearchSecretsByValue(context.Background(), client, "secret/app", "supersecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Matches["password"] != "supersecret" {
		t.Errorf("expected match on password, got %v", results[0].Matches)
	}
}

func TestSearchSecretsByValue_NoMatch(t *testing.T) {
	srv := newMockSearchValueServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	results, err := SearchSecretsByValue(context.Background(), client, "secret/app", "notfound")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
