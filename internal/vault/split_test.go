package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	hashivault "github.com/hashicorp/vault/api"
)

func newMockSplitServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"DB_HOST": "localhost",
						"DB_PASS": "secret",
						"API_KEY": "abc123",
					},
				},
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
}

func newSplitClient(t *testing.T, server *httptest.Server) *hashivault.Client {
	t.Helper()
	cfg := hashivault.DefaultConfig()
	cfg.Address = server.URL
	c, err := hashivault.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestSplitSecrets_EmptyPath(t *testing.T) {
	server := newMockSplitServer(t)
	defer server.Close()
	client := newSplitClient(t, server)

	_, err := SplitSecrets(context.Background(), client, "", map[string]string{"DB_HOST": "secret/db"}, "", false)
	if err == nil || err.Error() != "source path must not be empty" {
		t.Errorf("expected empty path error, got %v", err)
	}
}

func TestSplitSecrets_EmptyMappingAndNoDefault(t *testing.T) {
	server := newMockSplitServer(t)
	defer server.Close()
	client := newSplitClient(t, server)

	_, err := SplitSecrets(context.Background(), client, "secret/src", map[string]string{}, "", false)
	if err == nil {
		t.Error("expected error for empty mapping with no default, got nil")
	}
}

func TestSplitSecrets_DryRunDoesNotWrite(t *testing.T) {
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			writes++
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{"DB_HOST": "localhost"},
			},
		})
	}))
	defer server.Close()
	client := newSplitClient(t, server)

	_, err := SplitSecrets(context.Background(), client, "secret/src", map[string]string{"DB_HOST": "secret/db"}, "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if writes != 0 {
		t.Errorf("expected 0 writes in dry-run mode, got %d", writes)
	}
}

func TestSplitSecrets_RoutesKeysToCorrectDest(t *testing.T) {
	server := newMockSplitServer(t)
	defer server.Close()
	client := newSplitClient(t, server)

	mapping := map[string]string{
		"DB_HOST": "secret/db",
		"DB_PASS": "secret/db",
		"API_KEY": "secret/api",
	}

	result, err := SplitSecrets(context.Background(), client, "secret/src", mapping, "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dbKeys := result["secret/db"]
	sort.Strings(dbKeys)
	if len(dbKeys) != 2 || dbKeys[0] != "DB_HOST" || dbKeys[1] != "DB_PASS" {
		t.Errorf("expected DB_HOST and DB_PASS in secret/db, got %v", dbKeys)
	}

	apiKeys := result["secret/api"]
	if len(apiKeys) != 1 || apiKeys[0] != "API_KEY" {
		t.Errorf("expected API_KEY in secret/api, got %v", apiKeys)
	}
}
