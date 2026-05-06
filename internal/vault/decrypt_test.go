package vault

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockDecryptServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/data/myapp":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"API_KEY": "vault:v1:abc123encrypted",
						"PLAIN":   "unchanged",
					},
				},
			})
		case "/v1/transit/decrypt/mykey":
			w.WriteHeader(http.StatusOK)
			plain := base64.StdEncoding.EncodeToString([]byte("supersecret"))
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"plaintext": plain},
			})
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func newDecryptClient(t *testing.T, srv *httptest.Server) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestDecryptSecrets_EmptyPath(t *testing.T) {
	srv := newMockDecryptServer(t)
	defer srv.Close()
	client := newDecryptClient(t, srv)

	_, err := DecryptSecrets(context.Background(), client, "", "mykey", []string{".*"}, false)
	if err == nil || err.Error() != "path must not be empty" {
		t.Fatalf("expected empty path error, got %v", err)
	}
}

func TestDecryptSecrets_EmptyTransitKey(t *testing.T) {
	srv := newMockDecryptServer(t)
	defer srv.Close()
	client := newDecryptClient(t, srv)

	_, err := DecryptSecrets(context.Background(), client, "secret/data/myapp", "", []string{".*"}, false)
	if err == nil || err.Error() != "transit key must not be empty" {
		t.Fatalf("expected transit key error, got %v", err)
	}
}

func TestDecryptSecrets_NoPatterns(t *testing.T) {
	srv := newMockDecryptServer(t)
	defer srv.Close()
	client := newDecryptClient(t, srv)

	_, err := DecryptSecrets(context.Background(), client, "secret/data/myapp", "mykey", nil, false)
	if err == nil {
		t.Fatal("expected error for no patterns")
	}
}

func TestDecryptSecrets_InvalidPattern(t *testing.T) {
	srv := newMockDecryptServer(t)
	defer srv.Close()
	client := newDecryptClient(t, srv)

	_, err := DecryptSecrets(context.Background(), client, "secret/data/myapp", "mykey", []string{"[invalid"}, false)
	if err == nil {
		t.Fatal("expected invalid pattern error")
	}
}

func TestDecryptSecrets_DryRunDoesNotWrite(t *testing.T) {
	wrote := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v1/secret/data/myapp" {
			wrote = true
		}
		switch r.URL.Path {
		case "/v1/secret/data/myapp":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"API_KEY": "vault:v1:abc"},
				},
			})
		case "/v1/transit/decrypt/mykey":
			plain := base64.StdEncoding.EncodeToString([]byte("decrypted"))
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"plaintext": plain},
			})
		}
	}))
	defer srv.Close()
	client := newDecryptClient(t, srv)

	res, err := DecryptSecrets(context.Background(), client, "secret/data/myapp", "mykey", []string{"API_KEY"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["API_KEY"] != "decrypted" {
		t.Errorf("expected decrypted value, got %q", res["API_KEY"])
	}
	if wrote {
		t.Error("dry-run should not write to Vault")
	}
}
