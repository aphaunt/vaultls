package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockNormalizeServer(data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func newNormalizeClient(t *testing.T, srv *httptest.Server) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestNormalizeSecrets_EmptyPath(t *testing.T) {
	srv := newMockNormalizeServer(map[string]interface{}{})
	defer srv.Close()
	client := newNormalizeClient(t, srv)

	_, err := NormalizeSecrets(context.Background(), client, "", NormalizeOptions{})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestNormalizeSecrets_MutuallyExclusiveKeys(t *testing.T) {
	srv := newMockNormalizeServer(map[string]interface{}{"KEY": "val"})
	defer srv.Close()
	client := newNormalizeClient(t, srv)

	_, err := NormalizeSecrets(context.Background(), client, "secret/data/test", NormalizeOptions{
		LowerKeys: true,
		UpperKeys: true,
	})
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
}

func TestNormalizeSecrets_LowerKeys(t *testing.T) {
	srv := newMockNormalizeServer(map[string]interface{}{"MY_KEY": "hello", "ANOTHER": "world"})
	defer srv.Close()
	client := newNormalizeClient(t, srv)

	result, err := NormalizeSecrets(context.Background(), client, "secret/data/test", NormalizeOptions{
		LowerKeys: true,
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result["my_key"]; !ok {
		t.Errorf("expected key 'my_key', got %v", result)
	}
	if _, ok := result["another"]; !ok {
		t.Errorf("expected key 'another', got %v", result)
	}
}

func TestNormalizeSecrets_TrimSpace(t *testing.T) {
	srv := newMockNormalizeServer(map[string]interface{}{"key": "  spaced  "})
	defer srv.Close()
	client := newNormalizeClient(t, srv)

	result, err := NormalizeSecrets(context.Background(), client, "secret/data/test", NormalizeOptions{
		TrimSpace: true,
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["key"] != "spaced" {
		t.Errorf("expected trimmed value 'spaced', got %q", result["key"])
	}
}

func TestNormalizeSecrets_DryRunDoesNotWrite(t *testing.T) {
	writeCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writeCalled = true
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": map[string]interface{}{"K": "v"}}})
	}))
	defer srv.Close()
	client := newNormalizeClient(t, srv)

	_, err := NormalizeSecrets(context.Background(), client, "secret/data/test", NormalizeOptions{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if writeCalled {
		t.Error("expected no write in dry-run mode")
	}
}
