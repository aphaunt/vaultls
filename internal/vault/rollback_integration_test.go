package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/vault/api"
)

type rollbackState struct {
	mu      sync.Mutex
	writes  int
	lastData map[string]interface{}
}

func newStatefulRollbackServer(t *testing.T, state *rollbackState) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.RawQuery, "version="):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data":     map[string]interface{}{"token": "v1-secret"},
					"metadata": map[string]interface{}{"version": 1},
				},
			})
		case r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data":     map[string]interface{}{"token": "v2-secret"},
					"metadata": map[string]interface{}{"version": 2},
				},
			})
		case r.Method == http.MethodPost || r.Method == http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			state.mu.Lock()
			state.writes++
			if d, ok := body["data"].(map[string]interface{}); ok {
				state.lastData = d
			}
			state.mu.Unlock()
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestRollbackIntegration_FullRoundTrip(t *testing.T) {
	state := &rollbackState{}
	server := newStatefulRollbackServer(t, state)
	defer server.Close()

	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.SetToken("test-token")

	result, err := RollbackSecret(context.Background(), client, "secret/myapp/config", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success=true")
	}
	if result.Version != 1 {
		t.Errorf("expected version=1, got %d", result.Version)
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.writes != 1 {
		t.Errorf("expected 1 write, got %d", state.writes)
	}
}
