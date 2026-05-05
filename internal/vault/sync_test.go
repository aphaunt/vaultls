package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockSyncServer(t *testing.T, srcData, dstData map[string]interface{}) *httptest.Server {
	t.Helper()
	store := map[string]map[string]interface{}{
		"secret/data/src": {"data": srcData},
		"secret/data/dst": {"data": dstData},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if d, ok := store[r.URL.Path[1:]]; ok {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{"data": d})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case http.MethodPut, http.MethodPost:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			store[r.URL.Path[1:]] = body
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
}

func newSyncClient(t *testing.T, server *httptest.Server) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestSyncSecrets_EmptyPaths(t *testing.T) {
	server := newMockSyncServer(t, nil, nil)
	defer server.Close()
	client := newSyncClient(t, server)

	_, err := SyncSecrets(context.Background(), client, "", "dst", false, false)
	if err == nil {
		t.Fatal("expected error for empty src")
	}
}

func TestSyncSecrets_SamePath(t *testing.T) {
	server := newMockSyncServer(t, nil, nil)
	defer server.Close()
	client := newSyncClient(t, server)

	_, err := SyncSecrets(context.Background(), client, "env", "env", false, false)
	if err == nil {
		t.Fatal("expected error when src == dst")
	}
}

func TestSyncSecrets_DryRunDoesNotWrite(t *testing.T) {
	src := map[string]interface{}{"KEY": "value"}
	dst := map[string]interface{}{}
	server := newMockSyncServer(t, src, dst)
	defer server.Close()
	client := newSyncClient(t, server)

	results, err := SyncSecrets(context.Background(), client, "src", "dst", false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Action != "created" {
		t.Errorf("expected 1 created result, got %+v", results)
	}
}

func TestSyncSecrets_SkipsExistingWithoutOverwrite(t *testing.T) {
	src := map[string]interface{}{"KEY": "new"}
	dst := map[string]interface{}{"KEY": "old"}
	server := newMockSyncServer(t, src, dst)
	defer server.Close()
	client := newSyncClient(t, server)

	results, err := SyncSecrets(context.Background(), client, "src", "dst", false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Action != "skipped" {
		t.Errorf("expected skipped, got %+v", results)
	}
}
