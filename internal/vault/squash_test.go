package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockSquashServer(t *testing.T, reads map[string]map[string]interface{}, written *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if r.Method == http.MethodGet {
			data, ok := reads[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
			return
		}
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			*written = body
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
}

func newSquashClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestSquashSecrets_EmptyPaths(t *testing.T) {
	var written map[string]interface{}
	srv := newMockSquashServer(t, nil, &written)
	defer srv.Close()
	client := newSquashClient(t, srv)
	_, err := SquashSecrets(context.Background(), client, SquashOptions{Dest: "secret/data/out"})
	if err == nil || err.Error() != "at least one source path is required" {
		t.Fatalf("expected empty paths error, got: %v", err)
	}
}

func TestSquashSecrets_SamePath(t *testing.T) {
	var written map[string]interface{}
	srv := newMockSquashServer(t, nil, &written)
	defer srv.Close()
	client := newSquashClient(t, srv)
	_, err := SquashSecrets(context.Background(), client, SquashOptions{
		Paths: []string{"secret/data/a"},
		Dest:  "secret/data/a",
	})
	if err == nil {
		t.Fatal("expected same-path error")
	}
}

func TestSquashSecrets_DryRunDoesNotWrite(t *testing.T) {
	var written map[string]interface{}
	reads := map[string]map[string]interface{}{
		"/secret/data/a": {"foo": "bar"},
	}
	srv := newMockSquashServer(t, reads, &written)
	defer srv.Close()
	client := newSquashClient(t, srv)
	res, err := SquashSecrets(context.Background(), client, SquashOptions{
		Paths:  []string{"secret/data/a"},
		Dest:   "secret/data/out",
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", res["foo"])
	}
	if written != nil {
		t.Error("expected no write in dry-run mode")
	}
}

func TestSquashSecrets_MergesMultiplePaths(t *testing.T) {
	var written map[string]interface{}
	reads := map[string]map[string]interface{}{
		"/secret/data/a": {"key1": "val1"},
		"/secret/data/b": {"key2": "val2"},
	}
	srv := newMockSquashServer(t, reads, &written)
	defer srv.Close()
	client := newSquashClient(t, srv)
	res, err := SquashSecrets(context.Background(), client, SquashOptions{
		Paths: []string{"secret/data/a", "secret/data/b"},
		Dest:  "secret/data/out",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["key1"] != "val1" || res["key2"] != "val2" {
		t.Errorf("unexpected merged result: %v", res)
	}
}
