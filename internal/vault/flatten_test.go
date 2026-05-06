package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockFlattenServer(t *testing.T, initial map[string]interface{}) (*httptest.Server, *[]map[string]interface{}) {
	t.Helper()
	writes := &[]map[string]interface{}{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": initial}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			*writes = append(*writes, body)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
	return server, writes
}

func newFlattenClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestFlattenSecrets_EmptyPath(t *testing.T) {
	_, err := FlattenSecrets(context.Background(), newFlattenClient(t, "http://127.0.0.1"), "", "__", false)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestFlattenSecrets_DryRunDoesNotWrite(t *testing.T) {
	initial := map[string]interface{}{"db__host": "localhost", "db__port": "5432"}
	server, writes := newMockFlattenServer(t, initial)
	defer server.Close()

	result, err := FlattenSecrets(context.Background(), newFlattenClient(t, server.URL), "secret/myapp", "__", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*writes) != 0 {
		t.Errorf("expected no writes in dry-run, got %d", len(*writes))
	}
	if len(result) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result))
	}
}

func TestFlattenSecrets_WritesBack(t *testing.T) {
	initial := map[string]interface{}{"host": "localhost", "port": "5432"}
	server, writes := newMockFlattenServer(t, initial)
	defer server.Close()

	_, err := FlattenSecrets(context.Background(), newFlattenClient(t, server.URL), "secret/myapp", "__", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*writes) != 1 {
		t.Errorf("expected 1 write, got %d", len(*writes))
	}
}

func TestFlatKeys_Sorted(t *testing.T) {
	flat := map[string]string{"z": "1", "a": "2", "m": "3"}
	keys := FlatKeys(flat)
	if keys[0] != "a" || keys[1] != "m" || keys[2] != "z" {
		t.Errorf("keys not sorted: %v", keys)
	}
}

func TestFlattenMap_NestedKeys(t *testing.T) {
	input := map[string]interface{}{
		"db": map[string]interface{}{
			"host": "localhost",
			"port": "5432",
		},
		"app": "myapp",
	}
	out := make(map[string]string)
	flattenMap("", input, "__", out)
	if out["db__host"] != "localhost" {
		t.Errorf("expected db__host=localhost, got %s", out["db__host"])
	}
	if out["db__port"] != "5432" {
		t.Errorf("expected db__port=5432, got %s", out["db__port"])
	}
	if out["app"] != "myapp" {
		t.Errorf("expected app=myapp, got %s", out["app"])
	}
}
