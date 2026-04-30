package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockGCServer(t *testing.T, initial map[string]interface{}) (*httptest.Server, *map[string]interface{}) {
	t.Helper()
	current := initial
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": current}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"]; ok {
				current = d.(map[string]interface{})
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	return server, &current
}

func newGCClient(t *testing.T, addr string) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestGCSecrets_EmptyPath(t *testing.T) {
	_, err := GCSecrets(context.Background(), nil, "", []string{"tmp"}, false)
	if err == nil || err.Error() != "path must not be empty" {
		t.Errorf("expected empty path error, got %v", err)
	}
}

func TestGCSecrets_NoPatterns(t *testing.T) {
	_, err := GCSecrets(context.Background(), nil, "secret/data/app", nil, false)
	if err == nil || err.Error() != "at least one pattern must be provided" {
		t.Errorf("expected no patterns error, got %v", err)
	}
}

func TestGCSecrets_DryRunDoesNotDelete(t *testing.T) {
	initial := map[string]interface{}{"tmp_key": "val", "real_key": "keep"}
	server, current := newMockGCServer(t, initial)
	defer server.Close()

	client := newGCClient(t, server.URL)
	results, err := GCSecrets(context.Background(), client, "secret/data/app", []string{"tmp_"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Deleted {
		t.Error("expected Deleted=false in dry run")
	}
	if _, exists := (*current)["tmp_key"]; !exists {
		t.Error("dry run should not modify data")
	}
}

func TestGCSecrets_DeletesMatchingKeys(t *testing.T) {
	initial := map[string]interface{}{"tmp_key": "val", "real_key": "keep"}
	server, _ := newMockGCServer(t, initial)
	defer server.Close()

	client := newGCClient(t, server.URL)
	results, err := GCSecrets(context.Background(), client, "secret/data/app", []string{"tmp_"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Deleted {
		t.Error("expected Deleted=true")
	}
}
