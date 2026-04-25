package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/your-org/vaultls/internal/vault"
)

func newStatefulCloneServer(t *testing.T) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	store := map[string]map[string]interface{}{
		"secret/data/prod/api":    {"key": "abc123"},
		"secret/data/prod/db":     {"password": "hunter2"},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("list") == "true" {
				var keys []string
				prefix := path + "/"
				for k := range store {
					if strings.HasPrefix(k, prefix) {
						keys = append(keys, strings.TrimPrefix(k, prefix))
					}
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{"keys": keys},
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

func TestCloneIntegration_FullRoundTrip(t *testing.T) {
	srv := newStatefulCloneServer(t)
	defer srv.Close()
	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}

	result, err := vault.CloneSecrets(context.Background(), client,
		"secret/data/prod", "secret/data/staging",
		vault.CloneOptions{Overwrite: true})
	if err != nil {
		t.Fatalf("CloneSecrets error: %v", err)
	}

	if len(result.Copied) != 2 {
		t.Errorf("expected 2 secrets cloned, got %d", len(result.Copied))
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected 0 skipped, got %d", len(result.Skipped))
	}
}
