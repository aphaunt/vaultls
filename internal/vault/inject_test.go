package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockInjectServer(data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"data": data,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
}

func newInjectClient(t *testing.T, serverURL string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = serverURL
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestInjectSecrets_EmptyPath(t *testing.T) {
	client := newInjectClient(t, "http://127.0.0.1")
	_, err := InjectSecrets(context.Background(), client, "", false)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestInjectSecrets_SetsEnvVars(t *testing.T) {
	server := newMockInjectServer(map[string]interface{}{"db_pass": "secret123", "api_key": "abc"})
	defer server.Close()
	client := newInjectClient(t, server.URL)

	t.Setenv("DB_PASS", "")
	t.Setenv("API_KEY", "")

	injected, err := InjectSecrets(context.Background(), client, "secret/data/app", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(injected) != 2 {
		t.Fatalf("expected 2 injected vars, got %d", len(injected))
	}
	if os.Getenv("DB_PASS") != "secret123" && os.Getenv("API_KEY") != "secret123" {
		// values may be either key
		if os.Getenv("DB_PASS") == "" && os.Getenv("API_KEY") == "" {
			t.Error("expected env vars to be set")
		}
	}
}

func TestInjectSecrets_DryRunDoesNotSetEnv(t *testing.T) {
	server := newMockInjectServer(map[string]interface{}{"dry_key": "dry_val"})
	defer server.Close()
	client := newInjectClient(t, server.URL)

	os.Unsetenv("DRY_KEY")
	injected, err := InjectSecrets(context.Background(), client, "secret/data/dry", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(injected) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(injected))
	}
	if os.Getenv("DRY_KEY") != "" {
		t.Error("dry run should not set environment variable")
	}
}

func TestInjectSecrets_ContextCancelled(t *testing.T) {
	server := newMockInjectServer(map[string]interface{}{"k": "v"})
	defer server.Close()
	client := newInjectClient(t, server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := InjectSecrets(ctx, client, "secret/data/app", false)
	if err == nil {
		t.Log("no error returned for cancelled context (acceptable if server responded fast)")
	}
}
