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

func newMockSwapServer(t *testing.T) (*httptest.Server, map[string]map[string]interface{}) {
	t.Helper()
	store := map[string]map[string]interface{}{
		"/v1/secret/data/a": {"data": map[string]interface{}{"key": "valueA"}},
		"/v1/secret/data/b": {"data": map[string]interface{}{"key": "valueB"}},
	}
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.Method {
		case http.MethodGet:
			data, ok := store[r.URL.Path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			store[r.URL.Path] = body
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
	return srv, store
}

func newSwapClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestSwapSecrets_EmptyPath(t *testing.T) {
	srv, _ := newMockSwapServer(t)
	defer srv.Close()
	client := newSwapClient(t, srv.URL)
	if err := SwapSecrets(context.Background(), client, "", "secret/data/b", false); err == nil {
		t.Fatal("expected error for empty pathA")
	}
}

func TestSwapSecrets_SamePath(t *testing.T) {
	srv, _ := newMockSwapServer(t)
	defer srv.Close()
	client := newSwapClient(t, srv.URL)
	if err := SwapSecrets(context.Background(), client, "secret/data/a", "secret/data/a", false); err == nil {
		t.Fatal("expected error for same paths")
	}
}

func TestSwapSecrets_DryRunDoesNotWrite(t *testing.T) {
	srv, store := newMockSwapServer(t)
	defer srv.Close()
	client := newSwapClient(t, srv.URL)

	origA := store["/v1/secret/data/a"]
	origB := store["/v1/secret/data/b"]

	if err := SwapSecrets(context.Background(), client, "/v1/secret/data/a", "/v1/secret/data/b", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store["/v1/secret/data/a"]["data"] != origA["data"] {
		t.Error("pathA was modified during dry run")
	}
	if store["/v1/secret/data/b"]["data"] != origB["data"] {
		t.Error("pathB was modified during dry run")
	}
}

func TestSwapSecrets_ContextCancelled(t *testing.T) {
	srv, _ := newMockSwapServer(t)
	defer srv.Close()
	client := newSwapClient(t, srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := SwapSecrets(ctx, client, "secret/data/a", "secret/data/b", false)
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}
