package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newStatefulSetServer(t *testing.T) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	store := map[string]interface{}{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": store},
			})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if data, ok := body["data"].(map[string]interface{}); ok {
				for k, v := range data {
					store[k] = v
				}
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestSetIntegration_FullRoundTrip(t *testing.T) {
	server := newStatefulSetServer(t)
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)
	client.SetToken("test")

	ctx := context.Background()
	pairs := map[string]string{"alpha": "one", "beta": "two"}
	results, err := SetSecrets(ctx, client, "secret/app", pairs, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Written {
			t.Errorf("expected key %q to be written", r.Key)
		}
	}
}

func TestSetIntegration_DryRunDoesNotPersist(t *testing.T) {
	writeCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writeCalled = true
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{}},
		})
	}))
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)
	client.SetToken("test")

	_, err := SetSecrets(context.Background(), client, "secret/app", map[string]string{"k": "v"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if writeCalled {
		t.Fatal("expected no write in dry-run mode")
	}
}
