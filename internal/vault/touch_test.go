package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/yourusername/vaultls/internal/vault"
)

func newMockTouchServer(t *testing.T, path string, payload map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": payload,
				},
			})
		case http.MethodPost, http.MethodPut:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": payload})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestTouchSecrets_Success(t *testing.T) {
	payload := map[string]interface{}{"API_KEY": "abc123", "DB_PASS": "secret"}
	server := newMockTouchServer(t, "secret/data/myapp", payload)
	defer server.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = server.URL
	client, _ := vaultapi.NewClient(cfg)
	client.SetToken("test-token")

	results, err := vault.TouchSecrets(context.Background(), client, "secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("expected key %q to be successful", r.Key)
		}
	}
}

func TestTouchSecrets_EmptyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = server.URL
	client, _ := vaultapi.NewClient(cfg)
	client.SetToken("test-token")

	_, err := vault.TouchSecrets(context.Background(), client, "secret/data/empty")
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
}

func TestTouchSecrets_ContextCancelled(t *testing.T) {
	server := newMockTouchServer(t, "secret/data/myapp", map[string]interface{}{"key": "val"})
	defer server.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = server.URL
	client, _ := vaultapi.NewClient(cfg)
	client.SetToken("test-token")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := vault.TouchSecrets(ctx, client, "secret/data/myapp")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
