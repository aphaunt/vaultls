package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockRevertServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/data/myapp":
			v := r.URL.Query().Get("version")
			if v == "2" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{"key": "old-value"},
					},
				})
				return
			}
			http.Error(w, "version not found", http.StatusNotFound)
		case r.Method == http.MethodPut && r.URL.Path == "/v1/secret/data/myapp":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
		default:
			http.NotFound(w, r)
		}
	}))
}

func newRevertClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestRevertSecrets_EmptyPath(t *testing.T) {
	srv := newMockRevertServer(t)
	defer srv.Close()
	client := newRevertClient(t, srv)

	_, err := RevertSecrets(context.Background(), client, "", 1, false)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestRevertSecrets_InvalidVersion(t *testing.T) {
	srv := newMockRevertServer(t)
	defer srv.Close()
	client := newRevertClient(t, srv)

	_, err := RevertSecrets(context.Background(), client, "secret/myapp", 0, false)
	if err == nil {
		t.Fatal("expected error for version < 1")
	}
}

func TestRevertSecrets_DryRun(t *testing.T) {
	srv := newMockRevertServer(t)
	defer srv.Close()
	client := newRevertClient(t, srv)

	data, err := RevertSecrets(context.Background(), client, "secret/myapp", 2, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["key"] != "old-value" {
		t.Errorf("expected old-value, got %v", data["key"])
	}
}

func TestRevertSecrets_NilClient(t *testing.T) {
	_, err := RevertSecrets(context.Background(), nil, "secret/myapp", 1, false)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}
