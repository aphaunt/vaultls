package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockProtectServer(t *testing.T) (*httptest.Server, *sync.Map) {
	t.Helper()
	store := &sync.Map{}
	store.Store("secret/myapp", map[string]interface{}{"API_KEY": "abc123"})

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch r.Method {
		case http.MethodGet:
			val, ok := store.Load(path)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": val})
		case http.MethodPut, http.MethodPost:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			store.Store(path, body)
			w.WriteHeader(http.StatusNoContent)
		}
	})
	return httptest.NewServer(mux), store
}

func newProtectClient(t *testing.T, addr string) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestProtectSecrets_EmptyPath(t *testing.T) {
	svr, _ := newMockProtectServer(t)
	defer svr.Close()
	client := newProtectClient(t, svr.URL)

	if err := ProtectSecrets(context.Background(), client, ""); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestProtectSecrets_Success(t *testing.T) {
	svr, store := newMockProtectServer(t)
	defer svr.Close()
	client := newProtectClient(t, svr.URL)

	if err := ProtectSecrets(context.Background(), client, "secret/myapp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, _ := store.Load("secret/myapp")
	data := val.(map[string]interface{})
	if data[protectMetaKey] != "true" {
		t.Errorf("expected protect marker to be set, got: %v", data[protectMetaKey])
	}
}

func TestIsProtected_ReturnsFalseWhenAbsent(t *testing.T) {
	svr, _ := newMockProtectServer(t)
	defer svr.Close()
	client := newProtectClient(t, svr.URL)

	protected, err := IsProtected(context.Background(), client, "secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if protected {
		t.Error("expected not protected")
	}
}

func TestUnprotectSecrets_RemovesMarker(t *testing.T) {
	svr, store := newMockProtectServer(t)
	defer svr.Close()
	client := newProtectClient(t, svr.URL)

	store.Store("secret/myapp", map[string]interface{}{"API_KEY": "abc123", protectMetaKey: "true"})

	if err := UnprotectSecrets(context.Background(), client, "secret/myapp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, _ := store.Load("secret/myapp")
	data := val.(map[string]interface{})
	if _, exists := data[protectMetaKey]; exists {
		t.Error("expected protect marker to be removed")
	}
}
