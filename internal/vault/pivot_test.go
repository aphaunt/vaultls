package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockPivotServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/services":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []interface{}{"svcA", "svcB"},
				},
			})
		case "/v1/secret/data/services/svcA":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"env": "production",
						"port": "8080",
					},
				},
			})
		case "/v1/secret/data/services/svcB":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"env": "staging",
						"port": "9090",
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func newPivotClient(t *testing.T, srv *httptest.Server) *api.Client {
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

func TestPivotSecrets_EmptyPath(t *testing.T) {
	srv := newMockPivotServer(t)
	defer srv.Close()
	client := newPivotClient(t, srv)

	_, err := PivotSecrets(context.Background(), client, "", "env")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestPivotSecrets_EmptyPivotKey(t *testing.T) {
	srv := newMockPivotServer(t)
	defer srv.Close()
	client := newPivotClient(t, srv)

	_, err := PivotSecrets(context.Background(), client, "secret/services", "")
	if err == nil {
		t.Fatal("expected error for empty pivot key")
	}
}

func TestPivotSecrets_GroupsByPivotKey(t *testing.T) {
	srv := newMockPivotServer(t)
	defer srv.Close()
	client := newPivotClient(t, srv)

	results, err := PivotSecrets(context.Background(), client, "secret/services", "env")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	keys := map[string]bool{}
	for _, r := range results {
		keys[r.Key] = true
		if _, ok := r.Value["port"]; !ok {
			t.Errorf("expected 'port' in result for key %s", r.Key)
		}
		if _, ok := r.Value["env"]; ok {
			t.Errorf("pivot key 'env' should not appear in value map for %s", r.Key)
		}
	}
	if !keys["production"] || !keys["staging"] {
		t.Errorf("expected pivot keys 'production' and 'staging', got %v", keys)
	}
}
