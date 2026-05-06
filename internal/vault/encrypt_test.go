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

func newMockEncryptServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/secret/data/myapp"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"DB_PASS": "supersecret",
						"APP_NAME": "myapp",
					},
				},
			})
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "transit/encrypt"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"ciphertext": "vault:v1:abc123",
				},
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/secret/data/myapp"):
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newEncryptClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("creating vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestEncryptSecrets_EmptyPath(t *testing.T) {
	srv := newMockEncryptServer(t)
	defer srv.Close()
	c := newEncryptClient(t, srv)
	_, err := EncryptSecrets(context.Background(), c, "", "mykey", []string{"DB_*"}, false)
	if err == nil || !strings.Contains(err.Error(), "path must not be empty") {
		t.Fatalf("expected path error, got %v", err)
	}
}

func TestEncryptSecrets_EmptyTransitKey(t *testing.T) {
	srv := newMockEncryptServer(t)
	defer srv.Close()
	c := newEncryptClient(t, srv)
	_, err := EncryptSecrets(context.Background(), c, "secret/data/myapp", "", []string{"DB_*"}, false)
	if err == nil || !strings.Contains(err.Error(), "transit key must not be empty") {
		t.Fatalf("expected transit key error, got %v", err)
	}
}

func TestEncryptSecrets_NoPatterns(t *testing.T) {
	srv := newMockEncryptServer(t)
	defer srv.Close()
	c := newEncryptClient(t, srv)
	_, err := EncryptSecrets(context.Background(), c, "secret/data/myapp", "mykey", nil, false)
	if err == nil || !strings.Contains(err.Error(), "at least one key pattern") {
		t.Fatalf("expected pattern error, got %v", err)
	}
}

func TestEncryptSecrets_DryRunDoesNotWrite(t *testing.T) {
	writeCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/secret/data/myapp") {
			writeCalled = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"DB_PASS": "s3cr3t"}},
			})
			return
		}
		if strings.Contains(r.URL.Path, "transit/encrypt") {
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"ciphertext": "vault:v1:xyz"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	c := newEncryptClient(t, srv)
	_, err := EncryptSecrets(context.Background(), c, "secret/data/myapp", "mykey", []string{"DB_*"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if writeCalled {
		t.Fatal("expected no write on dry-run")
	}
}
