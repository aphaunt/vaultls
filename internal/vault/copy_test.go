package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/vaultls/internal/vault"
)

func newMockCopyServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := map[string]map[string]interface{}{
		"secret/data/src": {"KEY_A": "val_a", "KEY_B": "val_b"},
		"secret/data/dst": {"KEY_B": "old_b"},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/v1/"):]
		switch r.Method {
		case http.MethodGet:
			data, ok := store[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if inner, ok := body["data"].(map[string]interface{}); ok {
				store[path] = inner
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func TestCopySecrets_WithOverwrite(t *testing.T) {
	srv := newMockCopyServer(t)
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "test-token")
	result, err := vault.CopySecrets(context.Background(), client, "secret/data/src", "secret/data/dst", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.KeysCopied != 2 {
		t.Errorf("expected 2 keys copied, got %d", result.KeysCopied)
	}
}

func TestCopySecrets_WithoutOverwrite(t *testing.T) {
	srv := newMockCopyServer(t)
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "test-token")
	result, err := vault.CopySecrets(context.Background(), client, "secret/data/src", "secret/data/dst", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// KEY_B already exists at dst and overwrite=false, only KEY_A should be copied
	if result.KeysCopied != 1 {
		t.Errorf("expected 1 key copied, got %d", result.KeysCopied)
	}
}
