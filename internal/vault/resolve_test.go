package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockResolveServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/data/app":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"DB_HOST": "localhost",
						"DB_PORT": "5432",
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func newResolveClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("api.NewClient: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestResolveSecrets_Success(t *testing.T) {
	srv := newMockResolveServer(t)
	defer srv.Close()
	client := newResolveClient(t, srv.URL)

	results, err := ResolveSecrets(context.Background(), client, []string{"secret/data/app#DB_HOST"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Value != "localhost" {
		t.Fatalf("expected value 'localhost', got %+v", results)
	}
}

func TestResolveSecrets_MissingKey(t *testing.T) {
	srv := newMockResolveServer(t)
	defer srv.Close()
	client := newResolveClient(t, srv.URL)

	_, err := ResolveSecrets(context.Background(), client, []string{"secret/data/app#MISSING"})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestResolveSecrets_InvalidRef(t *testing.T) {
	srv := newMockResolveServer(t)
	defer srv.Close()
	client := newResolveClient(t, srv.URL)

	_, err := ResolveSecrets(context.Background(), client, []string{"secret/data/app"})
	if err == nil {
		t.Fatal("expected error for ref without '#key'")
	}
}

func TestResolveSecrets_EmptyRefs(t *testing.T) {
	srv := newMockResolveServer(t)
	defer srv.Close()
	client := newResolveClient(t, srv.URL)

	_, err := ResolveSecrets(context.Background(), client, []string{})
	if err == nil {
		t.Fatal("expected error for empty refs")
	}
}
