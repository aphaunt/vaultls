package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockRollbackServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "?version="):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"key": "oldvalue"},
					"metadata": map[string]interface{}{"version": 1},
				},
			})
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/v1/secret/data/"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"key": "oldvalue"},
					"metadata": map[string]interface{}{"version": 1},
				},
			})
		case r.Method == http.MethodPost || r.Method == http.MethodPut:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestRollbackSecret_InvalidPath(t *testing.T) {
	client, _ := api.NewClient(api.DefaultConfig())
	_, err := RollbackSecret(context.Background(), client, "", 1)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestRollbackSecret_InvalidVersion(t *testing.T) {
	client, _ := api.NewClient(api.DefaultConfig())
	_, err := RollbackSecret(context.Background(), client, "secret/foo", 0)
	if err == nil {
		t.Fatal("expected error for version < 1")
	}
}

func TestSplitMount(t *testing.T) {
	mount, path := splitMount("secret/foo/bar")
	if mount != "secret" || path != "foo/bar" {
		t.Errorf("unexpected split: mount=%q path=%q", mount, path)
	}
}

func TestSplitMount_NoSlash(t *testing.T) {
	mount, path := splitMount("secret")
	if mount != "secret" || path != "" {
		t.Errorf("unexpected split: mount=%q path=%q", mount, path)
	}
}
