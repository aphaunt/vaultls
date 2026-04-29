package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockPruneServer(t *testing.T, initial map[string]interface{}) (*httptest.Server, *map[string]interface{}) {
	t.Helper()
	store := initial
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": store}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"]; ok {
				store = d.(map[string]interface{})
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	return ts, &store
}

func newPruneClient(t *testing.T, addr string) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestPruneSecrets_EmptyPath(t *testing.T) {
	_, err := PruneSecrets(context.Background(), nil, "", nil, false)
	if err == nil || err.Error() != "path must not be empty" {
		t.Fatalf("expected empty path error, got %v", err)
	}
}

func TestPruneSecrets_RemovesEmptyValues(t *testing.T) {
	ts, _ := newMockPruneServer(t, map[string]interface{}{
		"KEY_A": "value",
		"KEY_B": "",
		"KEY_C": "   ",
	})
	defer ts.Close()
	client := newPruneClient(t, ts.URL)

	result, err := PruneSecrets(context.Background(), client, "secret/data/myapp", nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Deleted) != 2 {
		t.Errorf("expected 2 deleted, got %d: %v", len(result.Deleted), result.Deleted)
	}
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(result.Skipped))
	}
}

func TestPruneSecrets_DryRunDoesNotWrite(t *testing.T) {
	writeCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writeCalled = true
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": map[string]interface{}{"X": ""}}})
	}))
	defer ts.Close()
	client := newPruneClient(t, ts.URL)

	result, err := PruneSecrets(context.Background(), client, "secret/data/myapp", nil, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if writeCalled {
		t.Error("expected no write in dry-run mode")
	}
	if !result.DryRun {
		t.Error("expected DryRun to be true")
	}
}

func TestPruneSecrets_PatternMatch(t *testing.T) {
	ts, _ := newMockPruneServer(t, map[string]interface{}{
		"DB_PASSWORD": "secret123",
		"API_KEY":     "abc",
		"APP_NAME":    "myapp",
	})
	defer ts.Close()
	client := newPruneClient(t, ts.URL)

	result, err := PruneSecrets(context.Background(), client, "secret/data/myapp", []string{"PASSWORD", "KEY"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Deleted) != 2 {
		t.Errorf("expected 2 deleted, got %d: %v", len(result.Deleted), result.Deleted)
	}
}
