package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockRedactServer(t *testing.T, initial map[string]interface{}) (*httptest.Server, *map[string]interface{}) {
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
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	return ts, &store
}

func newRedactClient(t *testing.T, addr string) *vaultapi.Client {
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

func TestRedactSecrets_EmptyPath(t *testing.T) {
	_, err := RedactSecrets(context.Background(), nil, "", RedactOptions{Patterns: []string{".*"}})
	if err == nil || err.Error() != "path must not be empty" {
		t.Fatalf("expected empty path error, got %v", err)
	}
}

func TestRedactSecrets_NoPatterns(t *testing.T) {
	_, err := RedactSecrets(context.Background(), nil, "secret/data/test", RedactOptions{})
	if err == nil {
		t.Fatal("expected error for missing patterns")
	}
}

func TestRedactSecrets_InvalidPattern(t *testing.T) {
	_, err := RedactSecrets(context.Background(), nil, "secret/data/test", RedactOptions{Patterns: []string{"[invalid"}})
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestRedactSecrets_DryRunDoesNotWrite(t *testing.T) {
	initial := map[string]interface{}{"password": "s3cr3t", "username": "admin"}
	ts, store := newMockRedactServer(t, initial)
	defer ts.Close()

	client := newRedactClient(t, ts.URL)
	result, err := RedactSecrets(context.Background(), client, "secret/test", RedactOptions{
		Patterns: []string{"password"},
		DryRun:   true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["password"] != "***REDACTED***" {
		t.Errorf("expected password to be redacted, got %q", result["password"])
	}
	if result["username"] != "admin" {
		t.Errorf("expected username to be unchanged, got %q", result["username"])
	}
	if (*store)["password"] != "s3cr3t" {
		t.Error("dry run should not have written to store")
	}
}

func TestRedactSecrets_WritesOnNonDryRun(t *testing.T) {
	initial := map[string]interface{}{"api_key": "abc123", "host": "localhost"}
	ts, store := newMockRedactServer(t, initial)
	defer ts.Close()

	client := newRedactClient(t, ts.URL)
	_, err := RedactSecrets(context.Background(), client, "secret/test", RedactOptions{
		Patterns:    []string{"api_key"},
		Placeholder: "<hidden>",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*store)["api_key"] != "<hidden>" {
		t.Errorf("expected api_key to be redacted in store, got %v", (*store)["api_key"])
	}
}
