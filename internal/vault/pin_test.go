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

func newMockPinServer(t *testing.T) (*httptest.Server, *sync.Map) {
	t.Helper()
	store := &sync.Map{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch r.Method {
		case http.MethodGet:
			val, ok := store.Load(path)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": val}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"]; ok {
				store.Store(path, d)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	})
	return httptest.NewServer(mux), store
}

func newPinClient(t *testing.T, addr string) *vaultapi.Client {
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

func TestPinSecret_EmptyPath(t *testing.T) {
	svr, _ := newMockPinServer(t)
	defer svr.Close()
	client := newPinClient(t, svr.URL)

	err := PinSecret(context.Background(), client, "", 1)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestPinSecret_InvalidVersion(t *testing.T) {
	svr, _ := newMockPinServer(t)
	defer svr.Close()
	client := newPinClient(t, svr.URL)

	err := PinSecret(context.Background(), client, "secret/data/app", 0)
	if err == nil {
		t.Fatal("expected error for version < 1")
	}
}

func TestGetPin_ReturnsZeroWhenAbsent(t *testing.T) {
	svr, store := newMockPinServer(t)
	defer svr.Close()
	client := newPinClient(t, svr.URL)

	store.Store("/secret/data/app", map[string]interface{}{"key": "value"})

	v, err := GetPin(context.Background(), client, "secret/data/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Errorf("expected 0, got %d", v)
	}
}

func TestUnpinSecret_EmptyPath(t *testing.T) {
	svr, _ := newMockPinServer(t)
	defer svr.Close()
	client := newPinClient(t, svr.URL)

	err := UnpinSecret(context.Background(), client, "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}
