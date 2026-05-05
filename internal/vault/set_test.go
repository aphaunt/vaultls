package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockSetServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := map[string]interface{}{"existing": "value"}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func newSetClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestSetSecrets_EmptyPath(t *testing.T) {
	server := newMockSetServer(t)
	defer server.Close()
	client := newSetClient(t, server.URL)
	_, err := SetSecrets(context.Background(), client, "", map[string]string{"k": "v"}, false)
	if err == nil || err.Error() != "path must not be empty" {
		t.Fatalf("expected empty path error, got %v", err)
	}
}

func TestSetSecrets_EmptyPairs(t *testing.T) {
	server := newMockSetServer(t)
	defer server.Close()
	client := newSetClient(t, server.URL)
	_, err := SetSecrets(context.Background(), client, "secret/myapp", map[string]string{}, false)
	if err == nil {
		t.Fatal("expected error for empty pairs")
	}
}

func TestSetSecrets_DryRun(t *testing.T) {
	server := newMockSetServer(t)
	defer server.Close()
	client := newSetClient(t, server.URL)
	results, err := SetSecrets(context.Background(), client, "secret/myapp", map[string]string{"newkey": "newval"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || !results[0].Skipped {
		t.Fatalf("expected skipped result in dry-run, got %+v", results)
	}
}

func TestSetSecrets_WritesValues(t *testing.T) {
	server := newMockSetServer(t)
	defer server.Close()
	client := newSetClient(t, server.URL)
	results, err := SetSecrets(context.Background(), client, "secret/myapp", map[string]string{"foo": "bar"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || !results[0].Written {
		t.Fatalf("expected written result, got %+v", results)
	}
}
