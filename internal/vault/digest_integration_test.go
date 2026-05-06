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

type statefulDigestStore struct {
	mu   sync.RWMutex
	data map[string]map[string]interface{}
}

func newStatefulDigestServer(t *testing.T) (*httptest.Server, *statefulDigestStore) {
	t.Helper()
	store := &statefulDigestStore{
		data: make(map[string]map[string]interface{}),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store.mu.RLock()
		defer store.mu.RUnlock()
		path := r.URL.Path
		d, ok := store.data[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": d},
		})
	}))
	return server, store
}

func TestDigestIntegration_SameDataSameDigest(t *testing.T) {
	server, store := newStatefulDigestServer(t)
	defer server.Close()

	payload := map[string]interface{}{"DB_HOST": "localhost", "DB_PORT": "5432"}
	store.mu.Lock()
	store.data["/v1/secret/data/app"] = payload
	store.data["/v1/secret/data/app-copy"] = payload
	store.mu.Unlock()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)
	client.SetToken("test")

	match, err := CompareDigests(context.Background(), client, "secret/app", "secret/app-copy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Error("expected digests to match for identical data")
	}
}

func TestDigestIntegration_DifferentDataDifferentDigest(t *testing.T) {
	server, store := newStatefulDigestServer(t)
	defer server.Close()

	store.mu.Lock()
	store.data["/v1/secret/data/app-v1"] = map[string]interface{}{"KEY": "old"}
	store.data["/v1/secret/data/app-v2"] = map[string]interface{}{"KEY": "new"}
	store.mu.Unlock()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, _ := api.NewClient(cfg)
	client.SetToken("test")

	match, err := CompareDigests(context.Background(), client, "secret/app-v1", "secret/app-v2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match {
		t.Error("expected digests to differ for different data")
	}
}
