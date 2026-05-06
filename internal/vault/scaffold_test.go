package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockScaffoldServer(t *testing.T, stored map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch r.Method {
		case http.MethodGet:
			if d, ok := stored[path]; ok {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{"data": d})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				stored[path] = d
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
}

func newScaffoldClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	c, err := api.NewClient(&api.Config{Address: addr})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestScaffoldSecrets_EmptyPath(t *testing.T) {
	srv := newMockScaffoldServer(t, map[string]map[string]interface{}{})
	defer srv.Close()
	client := newScaffoldClient(t, srv.URL)
	_, err := ScaffoldSecrets(context.Background(), client, ScaffoldOptions{Path: "", Keys: []string{"foo"}})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestScaffoldSecrets_NoKeys(t *testing.T) {
	srv := newMockScaffoldServer(t, map[string]map[string]interface{}{})
	defer srv.Close()
	client := newScaffoldClient(t, srv.URL)
	_, err := ScaffoldSecrets(context.Background(), client, ScaffoldOptions{Path: "secret/app", Keys: nil})
	if err == nil {
		t.Fatal("expected error for empty keys")
	}
}

func TestScaffoldSecrets_DryRunDoesNotWrite(t *testing.T) {
	stored := map[string]map[string]interface{}{}
	srv := newMockScaffoldServer(t, stored)
	defer srv.Close()
	client := newScaffoldClient(t, srv.URL)
	result, err := ScaffoldSecrets(context.Background(), client, ScaffoldOptions{
		Path:   "secret/app",
		Keys:   []string{"DB_HOST", "DB_PORT"},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(result))
	}
	if len(stored) != 0 {
		t.Fatal("dry run must not write to vault")
	}
}

func TestScaffoldSecrets_WritesDefaults(t *testing.T) {
	stored := map[string]map[string]interface{}{}
	srv := newMockScaffoldServer(t, stored)
	defer srv.Close()
	client := newScaffoldClient(t, srv.URL)
	_, err := ScaffoldSecrets(context.Background(), client, ScaffoldOptions{
		Path:     "secret/app",
		Keys:     []string{"ENV", "DEBUG"},
		Defaults: map[string]string{"ENV": "production"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stored) == 0 {
		t.Fatal("expected data to be written")
	}
}
