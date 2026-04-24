package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/user/vaultls/internal/vault"
)

func newStatefulMockServer(t *testing.T, initial map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	store := initial
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/v1/"):]
		mu.Lock()
		defer mu.Unlock()
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

func TestCopyIntegration_FullRoundTrip(t *testing.T) {
	initial := map[string]map[string]interface{}{
		"secret/data/alpha": {"DB_HOST": "localhost", "DB_PASS": "secret"},
	}
	srv := newStatefulMockServer(t, initial)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "root")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := vault.CopySecrets(context.Background(), client, "secret/data/alpha", "secret/data/beta", true)
	if err != nil {
		t.Fatalf("CopySecrets: %v", err)
	}
	if result.KeysCopied != 2 {
		t.Errorf("expected 2 keys, got %d", result.KeysCopied)
	}
	if result.Destination != "secret/data/beta" {
		t.Errorf("unexpected destination: %s", result.Destination)
	}
}
