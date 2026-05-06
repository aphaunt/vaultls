package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockTruncateServer(data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				for k, v := range d {
					data[k] = v
				}
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
}

func newTruncateClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestTruncateSecrets_EmptyPath(t *testing.T) {
	srv := newMockTruncateServer(map[string]interface{}{})
	defer srv.Close()
	client := newTruncateClient(t, srv)
	_, err := TruncateSecrets(context.Background(), client, "", TruncateOptions{MaxLength: 10})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestTruncateSecrets_InvalidMaxLength(t *testing.T) {
	srv := newMockTruncateServer(map[string]interface{}{})
	defer srv.Close()
	client := newTruncateClient(t, srv)
	_, err := TruncateSecrets(context.Background(), client, "secret/data/test", TruncateOptions{MaxLength: 0})
	if err == nil {
		t.Fatal("expected error for zero max-length")
	}
}

func TestTruncateSecrets_TruncatesLongValues(t *testing.T) {
	data := map[string]interface{}{"key": "hello world this is a long value"}
	srv := newMockTruncateServer(data)
	defer srv.Close()
	client := newTruncateClient(t, srv)

	changed, err := TruncateSecrets(context.Background(), client, "secret/data/test", TruncateOptions{MaxLength: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changed) != 1 {
		t.Fatalf("expected 1 changed key, got %d", len(changed))
	}
	if changed["key"] != "hello..." {
		t.Errorf("expected 'hello...', got %q", changed["key"])
	}
}

func TestTruncateSecrets_DryRunDoesNotWrite(t *testing.T) {
	writeCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writeCalled = true
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": map[string]interface{}{"key": "a very long string value"}}})
	}))
	defer srv.Close()
	client := newTruncateClient(t, srv)

	_, err := TruncateSecrets(context.Background(), client, "secret/data/test", TruncateOptions{MaxLength: 5, DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if writeCalled {
		t.Error("dry-run should not write to vault")
	}
}

func TestTruncateSecrets_SkipsShortValues(t *testing.T) {
	data := map[string]interface{}{"key": "hi"}
	srv := newMockTruncateServer(data)
	defer srv.Close()
	client := newTruncateClient(t, srv)

	changed, err := TruncateSecrets(context.Background(), client, "secret/data/test", TruncateOptions{MaxLength: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changed) != 0 {
		t.Errorf("expected no changes, got %v", changed)
	}
}
