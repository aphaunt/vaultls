package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockSanitizeServer(t *testing.T, initial map[string]interface{}) (*httptest.Server, *map[string]interface{}) {
	t.Helper()
	store := initial
	ptr := &store
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": store}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				*ptr = d
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
	return ts, ptr
}

func newSanitizeClient(t *testing.T, addr string) *vaultapi.Client {
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

func TestSanitizeSecrets_EmptyPath(t *testing.T) {
	ts, _ := newMockSanitizeServer(t, map[string]interface{}{})
	defer ts.Close()
	client := newSanitizeClient(t, ts.URL)
	_, err := SanitizeSecrets(context.Background(), client, "", false)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSanitizeSecrets_TrimsWhitespace(t *testing.T) {
	initial := map[string]interface{}{"key": "  hello  ", "other": "clean"}
	ts, stored := newMockSanitizeServer(t, initial)
	defer ts.Close()
	client := newSanitizeClient(t, ts.URL)

	results, err := SanitizeSecrets(context.Background(), client, "secret/data/test", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	changed := 0
	for _, r := range results {
		if r.Changed {
			changed++
			if r.Key == "key" && r.NewValue != "hello" {
				t.Errorf("expected trimmed value 'hello', got %q", r.NewValue)
			}
		}
	}
	if changed != 1 {
		t.Errorf("expected 1 changed key, got %d", changed)
	}
	if (*stored)["key"] != "hello" {
		t.Errorf("expected stored value 'hello', got %v", (*stored)["key"])
	}
}

func TestSanitizeSecrets_DryRunDoesNotWrite(t *testing.T) {
	initial := map[string]interface{}{"key": "  value  "}
	ts, stored := newMockSanitizeServer(t, initial)
	defer ts.Close()
	client := newSanitizeClient(t, ts.URL)

	_, err := SanitizeSecrets(context.Background(), client, "secret/data/test", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*stored)["key"] != "  value  " {
		t.Errorf("dry run should not modify stored data, got %v", (*stored)["key"])
	}
}

func TestSanitizeSecrets_NoChanges(t *testing.T) {
	initial := map[string]interface{}{"key": "clean"}
	ts, _ := newMockSanitizeServer(t, initial)
	defer ts.Close()
	client := newSanitizeClient(t, ts.URL)

	results, err := SanitizeSecrets(context.Background(), client, "secret/data/test", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.Changed {
			t.Errorf("expected no changes, but key %q was marked changed", r.Key)
		}
	}
}
