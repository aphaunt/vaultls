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

func newMockMirrorServer(t *testing.T) (*httptest.Server, map[string]map[string]interface{}) {
	t.Helper()
	mu := &sync.Mutex{}
	store := map[string]map[string]interface{}{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		path := r.URL.Path
		switch r.Method {
		case http.MethodGet:
			data, ok := store[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				store[path] = d
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
	return ts, store
}

func newMirrorClient(t *testing.T, addr string) *api.Client {
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

func TestMirrorSecrets_EmptyPaths(t *testing.T) {
	ts, _ := newMockMirrorServer(t)
	defer ts.Close()
	client := newMirrorClient(t, ts.URL)

	_, err := MirrorSecrets(context.Background(), client, "", "dst", true, false)
	if err == nil {
		t.Fatal("expected error for empty src")
	}
	_, err = MirrorSecrets(context.Background(), client, "src", "", true, false)
	if err == nil {
		t.Fatal("expected error for empty dst")
	}
}

func TestMirrorSecrets_SamePath(t *testing.T) {
	ts, _ := newMockMirrorServer(t)
	defer ts.Close()
	client := newMirrorClient(t, ts.URL)

	_, err := MirrorSecrets(context.Background(), client, "secret/a", "secret/a", true, false)
	if err == nil {
		t.Fatal("expected error for identical src and dst")
	}
}

func TestMirrorSecrets_CopiesKeys(t *testing.T) {
	ts, store := newMockMirrorServer(t)
	defer ts.Close()
	client := newMirrorClient(t, ts.URL)
	store["/secret/src"] = map[string]interface{}{"foo": "bar", "baz": "qux"}

	result, err := MirrorSecrets(context.Background(), client, "/secret/src", "/secret/dst", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Copied) != 2 {
		t.Errorf("expected 2 copied, got %d", len(result.Copied))
	}
}

func TestMirrorSecrets_DeletesOrphans(t *testing.T) {
	ts, store := newMockMirrorServer(t)
	defer ts.Close()
	client := newMirrorClient(t, ts.URL)
	store["/secret/src"] = map[string]interface{}{"keep": "yes"}
	store["/secret/dst"] = map[string]interface{}{"keep": "yes", "orphan": "delete-me"}

	result, err := MirrorSecrets(context.Background(), client, "/secret/src", "/secret/dst", true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Deleted) != 1 || result.Deleted[0] != "orphan" {
		t.Errorf("expected orphan key deleted, got %v", result.Deleted)
	}
}

func TestMirrorSecrets_DryRunDoesNotWrite(t *testing.T) {
	ts, store := newMockMirrorServer(t)
	defer ts.Close()
	client := newMirrorClient(t, ts.URL)
	store["/secret/src"] = map[string]interface{}{"x": "1"}

	result, err := MirrorSecrets(context.Background(), client, "/secret/src", "/secret/dst", false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Copied) != 1 {
		t.Errorf("expected 1 reported as copied in dry-run")
	}
	if _, ok := store["/secret/dst"]; ok {
		t.Error("dry-run should not have written to dst")
	}
}
