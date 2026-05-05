package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockTrimServer(t *testing.T, initial map[string]interface{}, written *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": initial}})
		case http.MethodPut, http.MethodPost:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if written != nil {
				*written = body
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func newTrimClient(t *testing.T, serverURL string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = serverURL
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestTrimSecrets_EmptyPath(t *testing.T) {
	client := newTrimClient(t, "http://127.0.0.1")
	_, err := TrimSecrets(context.Background(), client, "", false)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestTrimSecrets_TrimsWhitespace(t *testing.T) {
	initial := map[string]interface{}{
		"key1": "  hello  ",
		"key2": "world",
		"key3": "\t spaced\n",
	}
	var written map[string]interface{}
	srv := newMockTrimServer(t, initial, &written)
	defer srv.Close()

	client := newTrimClient(t, srv.URL)
	results, err := TrimSecrets(context.Background(), client, "secret/data/test", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	changed := 0
	for _, r := range results {
		if r.Changed {
			changed++
		}
	}
	if changed != 2 {
		t.Errorf("expected 2 changed keys, got %d", changed)
	}
}

func TestTrimSecrets_DryRunDoesNotWrite(t *testing.T) {
	initial := map[string]interface{}{"key": "  value  "}
	var written map[string]interface{}
	srv := newMockTrimServer(t, initial, &written)
	defer srv.Close()

	client := newTrimClient(t, srv.URL)
	_, err := TrimSecrets(context.Background(), client, "secret/data/test", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if written != nil {
		t.Error("expected no write in dry-run mode")
	}
}

func TestTrimSecrets_NoChanges(t *testing.T) {
	initial := map[string]interface{}{"key": "clean"}
	srv := newMockTrimServer(t, initial, nil)
	defer srv.Close()

	client := newTrimClient(t, srv.URL)
	results, err := TrimSecrets(context.Background(), client, "secret/data/test", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Changed {
			t.Errorf("expected no changes, but key %q was marked changed", r.Key)
		}
	}
}
