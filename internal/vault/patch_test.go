package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockPatchServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := map[string]interface{}{
		"username": "alice",
		"password": "secret",
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": store}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				for k, v := range d {
					store[k] = v
				}
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
}

func newPatchClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	c.SetToken("test-token")
	return c
}

func TestPatchSecrets_EmptyPath(t *testing.T) {
	svr := newMockPatchServer(t)
	defer svr.Close()
	c := newPatchClient(t, svr.URL)
	_, err := PatchSecrets(context.Background(), c, "", map[string]string{"k": "v"}, false)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestPatchSecrets_EmptyUpdates(t *testing.T) {
	svr := newMockPatchServer(t)
	defer svr.Close()
	c := newPatchClient(t, svr.URL)
	_, err := PatchSecrets(context.Background(), c, "secret/data/app", map[string]string{}, false)
	if err == nil {
		t.Fatal("expected error for empty updates")
	}
}

func TestPatchSecrets_DryRunDoesNotWrite(t *testing.T) {
	svr := newMockPatchServer(t)
	defer svr.Close()
	c := newPatchClient(t, svr.URL)
	results, err := PatchSecrets(context.Background(), c, "secret/data/app", map[string]string{"password": "new"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Skipped {
		t.Error("expected result to be skipped in dry-run")
	}
}

func TestPatchSecrets_UpdatesExistingKey(t *testing.T) {
	svr := newMockPatchServer(t)
	defer svr.Close()
	c := newPatchClient(t, svr.URL)
	results, err := PatchSecrets(context.Background(), c, "secret/data/app", map[string]string{"username": "bob"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OldVal != "alice" {
		t.Errorf("expected old value 'alice', got %q", results[0].OldVal)
	}
	if results[0].NewVal != "bob" {
		t.Errorf("expected new value 'bob', got %q", results[0].NewVal)
	}
}
