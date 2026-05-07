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

func newMockRestoreServer(t *testing.T, written map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			written[r.URL.Path] = body
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
}

func newRestoreClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
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

func TestRestoreSecrets_EmptyDestPath(t *testing.T) {
	wrote := map[string]map[string]interface{}{}
	srv := newMockRestoreServer(t, wrote)
	defer srv.Close()
	c := newRestoreClient(t, srv)

	snap, _ := json.Marshal(map[string]string{"key": "val"})
	_, err := RestoreSecrets(context.Background(), c, "", string(snap), true, false)
	if err == nil || !strings.Contains(err.Error(), "destination path") {
		t.Fatalf("expected destination path error, got %v", err)
	}
}

func TestRestoreSecrets_EmptySnapshot(t *testing.T) {
	wrote := map[string]map[string]interface{}{}
	srv := newMockRestoreServer(t, wrote)
	defer srv.Close()
	c := newRestoreClient(t, srv)

	_, err := RestoreSecrets(context.Background(), c, "secret/myapp", "", true, false)
	if err == nil || !strings.Contains(err.Error(), "snapshot JSON") {
		t.Fatalf("expected snapshot JSON error, got %v", err)
	}
}

func TestRestoreSecrets_DryRunDoesNotWrite(t *testing.T) {
	wrote := map[string]map[string]interface{}{}
	srv := newMockRestoreServer(t, wrote)
	defer srv.Close()
	c := newRestoreClient(t, srv)

	snap, _ := json.Marshal(map[string]string{"alpha": "one"})
	restored, err := RestoreSecrets(context.Background(), c, "secret/myapp", string(snap), true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wrote) != 0 {
		t.Fatalf("expected no writes in dry-run, got %d", len(wrote))
	}
	if len(restored) != 1 {
		t.Fatalf("expected 1 restored entry, got %d", len(restored))
	}
}

func TestRestoreSecrets_WritesKeys(t *testing.T) {
	wrote := map[string]map[string]interface{}{}
	srv := newMockRestoreServer(t, wrote)
	defer srv.Close()
	c := newRestoreClient(t, srv)

	snap, _ := json.Marshal(map[string]string{"beta": "two", "gamma": "three"})
	restored, err := RestoreSecrets(context.Background(), c, "secret/myapp", string(snap), true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(restored) != 2 {
		t.Fatalf("expected 2 restored entries, got %d", len(restored))
	}
	if len(wrote) != 2 {
		t.Fatalf("expected 2 vault writes, got %d", len(wrote))
	}
}
