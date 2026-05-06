package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockDedupeServer(data map[string]interface{}) (*httptest.Server, *[]map[string]interface{}) {
	writes := &[]map[string]interface{}{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": data},
			})
			return
		}
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			*writes = append(*writes, body)
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	return server, writes
}

func newDedupeClient(t *testing.T, serverURL string) *api.Client {
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

func TestDedupeSecrets_EmptyPath(t *testing.T) {
	server, _ := newMockDedupeServer(nil)
	defer server.Close()
	client := newDedupeClient(t, server.URL)

	_, err := DedupeSecrets(context.Background(), client, "", false)
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}

func TestDedupeSecrets_NoDuplicates(t *testing.T) {
	data := map[string]interface{}{"alpha": "foo", "beta": "bar", "gamma": "baz"}
	server, writes := newMockDedupeServer(data)
	defer server.Close()
	client := newDedupeClient(t, server.URL)

	results, err := DedupeSecrets(context.Background(), client, "myapp/config", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no duplicates, got %d", len(results))
	}
	if len(*writes) != 0 {
		t.Error("expected no writes when no duplicates found")
	}
}

func TestDedupeSecrets_DetectsDuplicates(t *testing.T) {
	data := map[string]interface{}{"key_a": "secret", "key_b": "secret", "key_c": "unique"}
	server, _ := newMockDedupeServer(data)
	defer server.Close()
	client := newDedupeClient(t, server.URL)

	results, err := DedupeSecrets(context.Background(), client, "myapp/config", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(results))
	}
	if results[0].Key != "key_a" {
		t.Errorf("expected canonical key to be key_a, got %s", results[0].Key)
	}
	if len(results[0].Duplicates) != 1 || results[0].Duplicates[0] != "key_b" {
		t.Errorf("unexpected duplicates: %v", results[0].Duplicates)
	}
}

func TestDedupeSecrets_DryRunDoesNotWrite(t *testing.T) {
	data := map[string]interface{}{"x": "val", "y": "val"}
	server, writes := newMockDedupeServer(data)
	defer server.Close()
	client := newDedupeClient(t, server.URL)

	results, err := DedupeSecrets(context.Background(), client, "myapp/config", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected duplicates to be detected in dry-run")
	}
	if len(*writes) != 0 {
		t.Error("expected no writes in dry-run mode")
	}
}
