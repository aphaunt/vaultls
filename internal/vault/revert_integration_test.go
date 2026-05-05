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

type statefulRevertStore struct {
	mu       sync.Mutex
	versions []map[string]interface{}
	current  map[string]interface{}
}

func newStatefulRevertServer(t *testing.T, initial map[string]interface{}) (*httptest.Server, *statefulRevertStore) {
	t.Helper()
	store := &statefulRevertStore{
		versions: []map[string]interface{}{initial},
		current:  initial,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store.mu.Lock()
		defer store.mu.Unlock()
		switch {
		case r.Method == http.MethodGet:
			v := r.URL.Query().Get("version")
			idx := len(store.versions) - 1
			if v != "" {
				var vi int
				json.Unmarshal([]byte(v), &vi)
				idx = vi - 1
			}
			if idx < 0 || idx >= len(store.versions) {
				http.NotFound(w, r)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": store.versions[idx]},
			})
		case r.Method == http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				store.versions = append(store.versions, d)
				store.current = d
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		default:
			http.NotFound(w, r)
		}
	}))
	return srv, store
}

func TestRevertIntegration_FullRoundTrip(t *testing.T) {
	initial := map[string]interface{}{"token": "abc123"}
	srv, store := newStatefulRevertServer(t, initial)
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, _ := vaultapi.NewClient(cfg)
	client.SetToken("test")

	// Revert to version 1 (dry-run)
	data, err := RevertSecrets(context.Background(), client, "secret/myapp", 1, true)
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	if data["token"] != "abc123" {
		t.Errorf("expected abc123, got %v", data["token"])
	}
	if len(store.versions) != 1 {
		t.Error("dry-run should not write a new version")
	}

	// Revert for real
	_, err = RevertSecrets(context.Background(), client, "secret/myapp", 1, false)
	if err != nil {
		t.Fatalf("revert failed: %v", err)
	}
	if len(store.versions) != 2 {
		t.Errorf("expected 2 versions after revert, got %d", len(store.versions))
	}
}
