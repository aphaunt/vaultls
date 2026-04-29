package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockArchiveServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "metadata/prod"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"api_key", "db_pass"},
				},
			})
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "data/prod"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"value": "secret123"},
				},
			})
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "data/archive"):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "metadata/prod"):
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestArchiveSecrets_EmptyPath(t *testing.T) {
	client, _ := vaultapi.NewClient(vaultapi.DefaultConfig())
	_, err := ArchiveSecrets(context.Background(), client, "", "archive", false, false)
	if err == nil || !strings.Contains(err.Error(), "source path") {
		t.Fatalf("expected source path error, got %v", err)
	}
}

func TestArchiveSecrets_SamePath(t *testing.T) {
	client, _ := vaultapi.NewClient(vaultapi.DefaultConfig())
	_, err := ArchiveSecrets(context.Background(), client, "prod", "prod", false, false)
	if err == nil || !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("expected differ error, got %v", err)
	}
}

func TestArchiveSecrets_DryRun(t *testing.T) {
	srv := newMockArchiveServer(t)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := vaultapi.NewClient(cfg)
	client.SetToken("test")

	results, err := ArchiveSecrets(context.Background(), client, "prod", "archive", false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if !r.Skipped || r.Reason != "dry-run" {
			t.Errorf("expected dry-run skip, got %+v", r)
		}
	}
}

func TestArchiveSecrets_WithDelete(t *testing.T) {
	srv := newMockArchiveServer(t)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := vaultapi.NewClient(cfg)
	client.SetToken("test")

	results, err := ArchiveSecrets(context.Background(), client, "prod", "archive", true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if !r.Archived {
			t.Errorf("expected archived=true, got %+v", r)
		}
	}
}
