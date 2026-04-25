package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/yourusername/vaultls/internal/vault"
)

func newMockRenameServer(t *testing.T) (*httptest.Server, *sync.Map) {
	t.Helper()
	store := &sync.Map{}

	// Seed source data
	store.Store("secret/data/src", map[string]interface{}{"key": "value"})

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch r.Method {
		case http.MethodGet:
			val, ok := store.Load(path[4:]) // strip /v1/
			if !ok {
				http.NotFound(w, r)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": val}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if data, ok := body["data"]; ok {
				store.Store(path[4:], data)
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			store.Delete(path[4:])
			w.WriteHeader(http.StatusNoContent)
		}
	})
	return httptest.NewServer(mux), store
}

func TestRenameSecrets_Success(t *testing.T) {
	server, store := newMockRenameServer(t)
	defer server.Close()

	client, err := vault.NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx := context.Background()
	if err := vault.RenameSecrets(ctx, client, "secret/data/src", "secret/data/dst", false); err != nil {
		t.Fatalf("RenameSecrets: %v", err)
	}

	_, srcExists := store.Load("secret/data/src")
	if srcExists {
		t.Error("expected source to be deleted after rename")
	}
	_, dstExists := store.Load("secret/data/dst")
	if !dstExists {
		t.Error("expected destination to exist after rename")
	}
}

func TestRenameSecrets_SamePath(t *testing.T) {
	client, _ := vault.NewClient("http://localhost", "token")
	err := vault.RenameSecrets(context.Background(), client, "secret/data/same", "secret/data/same", false)
	if err == nil {
		t.Fatal("expected error for same src and dst path")
	}
}

func TestRenameSecrets_DestinationExists_NoOverwrite(t *testing.T) {
	server, store := newMockRenameServer(t)
	defer server.Close()
	store.Store("secret/data/dst", map[string]interface{}{"existing": "data"})

	client, _ := vault.NewClient(server.URL, "test-token")
	err := vault.RenameSecrets(context.Background(), client, "secret/data/src", "secret/data/dst", false)
	if err == nil {
		t.Fatal("expected error when destination exists and overwrite is false")
	}
}
