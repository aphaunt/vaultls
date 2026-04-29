package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockFmtServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := map[string]interface{}{
		"Database_URL": "postgres://localhost",
		"Api_Key":      "secret123",
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": store})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				for k, v := range d {
					store[k] = v
				}
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func newFmtClient(t *testing.T, srv *httptest.Server) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestFmtSecrets_EmptyPath(t *testing.T) {
	srv := newMockFmtServer(t)
	defer srv.Close()
	c := newFmtClient(t, srv)
	_, err := FmtSecrets(context.Background(), c, "", "upper", false)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestFmtSecrets_InvalidStyle(t *testing.T) {
	srv := newMockFmtServer(t)
	defer srv.Close()
	c := newFmtClient(t, srv)
	_, err := FmtSecrets(context.Background(), c, "secret/data/app", "title", false)
	if err == nil {
		t.Fatal("expected error for invalid style")
	}
}

func TestFmtSecrets_UpperStyle(t *testing.T) {
	srv := newMockFmtServer(t)
	defer srv.Close()
	c := newFmtClient(t, srv)
	results, err := FmtSecrets(context.Background(), c, "secret/data/app", "upper", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if r.NewVal != strings.ToUpper(r.OldVal) {
			t.Errorf("expected upper-case key, got %q", r.NewVal)
		}
	}
}

func TestFmtSecrets_DryRunDoesNotWrite(t *testing.T) {
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writes++
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"KEY": "val"}})
	}))
	defer srv.Close()
	c := newFmtClient(t, srv)
	_, err := FmtSecrets(context.Background(), c, "secret/data/app", "lower", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if writes != 0 {
		t.Errorf("dry-run should not write, got %d writes", writes)
	}
}
