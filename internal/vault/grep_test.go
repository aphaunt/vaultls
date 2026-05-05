package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockGrepServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/myapp":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"config", "creds"}},
			})
		case "/v1/secret/data/myapp/config":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"DB_HOST": "localhost", "DB_PORT": "5432"}},
			})
		case "/v1/secret/data/myapp/creds":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"API_KEY": "secret123", "TOKEN": "abc"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func newGrepClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestGrepSecrets_EmptyPath(t *testing.T) {
	srv := newMockGrepServer(t)
	defer srv.Close()
	c := newGrepClient(t, srv)
	_, err := GrepSecrets(context.Background(), c, "", "DB", false)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestGrepSecrets_EmptyPattern(t *testing.T) {
	srv := newMockGrepServer(t)
	defer srv.Close()
	c := newGrepClient(t, srv)
	_, err := GrepSecrets(context.Background(), c, "secret/myapp", "", false)
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestGrepSecrets_MatchByKey(t *testing.T) {
	srv := newMockGrepServer(t)
	defer srv.Close()
	c := newGrepClient(t, srv)
	results, err := GrepSecrets(context.Background(), c, "secret/myapp", "DB_", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 matches, got %d", len(results))
	}
}

func TestGrepSecrets_MatchByValue(t *testing.T) {
	srv := newMockGrepServer(t)
	defer srv.Close()
	c := newGrepClient(t, srv)
	results, err := GrepSecrets(context.Background(), c, "secret/myapp", "secret123", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 match, got %d", len(results))
	}
}

func TestGrepSecrets_NoMatch(t *testing.T) {
	srv := newMockGrepServer(t)
	defer srv.Close()
	c := newGrepClient(t, srv)
	results, err := GrepSecrets(context.Background(), c, "secret/myapp", "NONEXISTENT", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 matches, got %d", len(results))
	}
}
