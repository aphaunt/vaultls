package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockExpireServer(customMeta map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"custom_metadata": customMeta,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
}

func newExpireClient(t *testing.T, server *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = server.URL
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestExpireSecrets_EmptyPath(t *testing.T) {
	server := newMockExpireServer(nil)
	defer server.Close()
	client := newExpireClient(t, server)

	_, err := ExpireSecrets(context.Background(), client, "", time.Now())
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestExpireSecrets_NoCustomMeta(t *testing.T) {
	server := newMockExpireServer(map[string]interface{}{})
	defer server.Close()
	client := newExpireClient(t, server)

	results, err := ExpireSecrets(context.Background(), client, "secret/myapp", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestExpireSecrets_DetectsExpired(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	server := newMockExpireServer(map[string]interface{}{
		"API_KEY": past,
	})
	defer server.Close()
	client := newExpireClient(t, server)

	results, err := ExpireSecrets(context.Background(), client, "secret/myapp", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Expired {
		t.Errorf("expected key to be expired")
	}
	if results[0].Key != "API_KEY" {
		t.Errorf("expected key API_KEY, got %s", results[0].Key)
	}
}

func TestExpireSecrets_DetectsValid(t *testing.T) {
	future := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	server := newMockExpireServer(map[string]interface{}{
		"DB_PASS": future,
	})
	defer server.Close()
	client := newExpireClient(t, server)

	results, err := ExpireSecrets(context.Background(), client, "secret/myapp", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Expired {
		t.Errorf("expected key to be valid (not expired)")
	}
	if results[0].TTL <= 0 {
		t.Errorf("expected positive TTL, got %v", results[0].TTL)
	}
}
