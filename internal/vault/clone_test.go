package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultls/internal/vault"
)

func newMockCloneServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := map[string]map[string]interface{}{
		"secret/data/src/db": {"password": "s3cr3t"},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("list") == "true" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{"keys": []string{"db"}},
				})
				return
			}
			if data, ok := store[path]; ok {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
				return
			}
			w.WriteHeader(http.StatusNotFound)
		case http.MethodPut, http.MethodPost:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			store[path] = body
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestCloneSecrets_Success(t *testing.T) {
	srv := newMockCloneServer(t)
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "test-token")
	result, err := vault.CloneSecrets(context.Background(), client, "secret/data/src", "secret/data/dst", vault.CloneOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Copied) != 1 || result.Copied[0] != "db" {
		t.Errorf("expected [db] copied, got %v", result.Copied)
	}
}

func TestCloneSecrets_SamePath(t *testing.T) {
	srv := newMockCloneServer(t)
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "test-token")
	_, err := vault.CloneSecrets(context.Background(), client, "secret/data/src", "secret/data/src", vault.CloneOptions{})
	if err == nil {
		t.Fatal("expected error for same src/dst path")
	}
}

func TestCloneSecrets_DryRun(t *testing.T) {
	srv := newMockCloneServer(t)
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "test-token")
	result, err := vault.CloneSecrets(context.Background(), client, "secret/data/src", "secret/data/dry", vault.CloneOptions{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Copied) != 1 {
		t.Errorf("expected 1 entry reported as copied (dry-run), got %d", len(result.Copied))
	}
}

func TestCloneSecrets_SkipsExistingWithoutOverwrite(t *testing.T) {
	srv := newMockCloneServer(t)
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "test-token")
	// First clone populates dst
	vault.CloneSecrets(context.Background(), client, "secret/data/src", "secret/data/dst2", vault.CloneOptions{})
	// Second clone without overwrite should skip
	result, err := vault.CloneSecrets(context.Background(), client, "secret/data/src", "secret/data/dst2", vault.CloneOptions{Overwrite: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Skipped) == 0 {
		t.Error("expected at least one skipped key")
	}
}
