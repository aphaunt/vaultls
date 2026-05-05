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

func newMockTraceServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/secret/metadata/myapp") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"versions": map[string]interface{}{
						"1": map[string]interface{}{
							"created_time":    "2024-01-01T00:00:00Z",
							"destruction_time": "",
						},
						"2": map[string]interface{}{
							"created_time":    "2024-02-01T00:00:00Z",
							"destruction_time": "2024-03-01T00:00:00Z",
						},
					},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))
}

func newTraceClient(t *testing.T, addr string) *vaultapi.Client {
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

func TestTraceSecrets_EmptyPath(t *testing.T) {
	svr := newMockTraceServer(t)
	defer svr.Close()
	client := newTraceClient(t, svr.URL)

	_, err := TraceSecrets(context.Background(), client, "")
	if err == nil || !strings.Contains(err.Error(), "must not be empty") {
		t.Fatalf("expected empty path error, got %v", err)
	}
}

func TestTraceSecrets_InvalidPath(t *testing.T) {
	svr := newMockTraceServer(t)
	defer svr.Close()
	client := newTraceClient(t, svr.URL)

	_, err := TraceSecrets(context.Background(), client, "noslash")
	if err == nil || !strings.Contains(err.Error(), "invalid path") {
		t.Fatalf("expected invalid path error, got %v", err)
	}
}

func TestTraceSecrets_Success(t *testing.T) {
	svr := newMockTraceServer(t)
	defer svr.Close()
	client := newTraceClient(t, svr.URL)

	result, err := TraceSecrets(context.Background(), client, "secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Path != "secret/myapp" {
		t.Errorf("expected path secret/myapp, got %s", result.Path)
	}
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
	var hasDeleted bool
	for _, e := range result.Entries {
		if e.Operation == "deleted" {
			hasDeleted = true
		}
	}
	if !hasDeleted {
		t.Error("expected at least one deleted entry")
	}
}
