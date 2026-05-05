package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockRotateServer(t *testing.T, initial map[string]interface{}) (*httptest.Server, *map[string]interface{}) {
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
			if d, ok := body["data"].(map[string]interface{}); ok {
				store = d
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
	return ts, &store
}

func newRotateClient(t *testing.T, addr string) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestRotateSecrets_EmptyPath(t *testing.T) {
	_, err := RotateSecrets(context.Background(), nil, "", nil, false)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected empty-path error, got %v", err)
	}
}

func TestRotateSecrets_NilGenerator(t *testing.T) {
	ts, _ := newMockRotateServer(t, map[string]interface{}{"k": "v"})
	defer ts.Close()
	client := newRotateClient(t, ts.URL)
	_, err := RotateSecrets(context.Background(), client, "secret/data/test", nil, false)
	if err == nil || !strings.Contains(err.Error(), "generator") {
		t.Fatalf("expected generator error, got %v", err)
	}
}

func TestRotateSecrets_DryRunDoesNotWrite(t *testing.T) {
	initial := map[string]interface{}{"api_key": "old-value"}
	ts, store := newMockRotateServer(t, initial)
	defer ts.Close()
	client := newRotateClient(t, ts.URL)

	gen := func(_, _ string) (string, error) { return "new-value", nil }
	results, err := RotateSecrets(context.Background(), client, "secret/test", gen, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Rotated {
		t.Error("expected Rotated=false in dry-run mode")
	}
	if (*store)["api_key"] != "old-value" {
		t.Error("store should not be mutated in dry-run mode")
	}
}

func TestRotateSecrets_WritesNewValues(t *testing.T) {
	initial := map[string]interface{}{"password": "hunter2"}
	ts, store := newMockRotateServer(t, initial)
	defer ts.Close()
	client := newRotateClient(t, ts.URL)

	gen := func(key, _ string) (string, error) { return "rotated-" + key, nil }
	results, err := RotateSecrets(context.Background(), client, "secret/test", gen, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || !results[0].Rotated {
		t.Fatal("expected one rotated result")
	}
	if (*store)["password"] != "rotated-password" {
		t.Errorf("store not updated, got %v", (*store)["password"])
	}
}
