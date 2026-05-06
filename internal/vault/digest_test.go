package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockDigestServer(data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"data": data,
			},
		}
		json.NewEncoder(w).Encode(payload)
	}))
}

func newDigestClient(t *testing.T, server *httptest.Server) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestDigestSecrets_EmptyPath(t *testing.T) {
	server := newMockDigestServer(map[string]interface{}{})
	defer server.Close()
	client := newDigestClient(t, server)

	_, err := DigestSecrets(context.Background(), client, "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestDigestSecrets_Success(t *testing.T) {
	server := newMockDigestServer(map[string]interface{}{
		"KEY_A": "value1",
		"KEY_B": "value2",
	})
	defer server.Close()
	client := newDigestClient(t, server)

	result, err := DigestSecrets(context.Background(), client, "secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Digest == "" {
		t.Error("expected non-empty digest")
	}
	if result.Keys != 2 {
		t.Errorf("expected 2 keys, got %d", result.Keys)
	}
}

func TestDigestSecrets_Deterministic(t *testing.T) {
	data := map[string]interface{}{"Z": "last", "A": "first", "M": "middle"}
	server := newMockDigestServer(data)
	defer server.Close()
	client := newDigestClient(t, server)

	r1, err := DigestSecrets(context.Background(), client, "secret/myapp")
	if err != nil {
		t.Fatalf("first digest error: %v", err)
	}
	r2, err := DigestSecrets(context.Background(), client, "secret/myapp")
	if err != nil {
		t.Fatalf("second digest error: %v", err)
	}
	if r1.Digest != r2.Digest {
		t.Errorf("digests differ: %s vs %s", r1.Digest, r2.Digest)
	}
}

func TestDigestSecrets_ContextCancelled(t *testing.T) {
	server := newMockDigestServer(map[string]interface{}{"k": "v"})
	defer server.Close()
	client := newDigestClient(t, server)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := DigestSecrets(ctx, client, "secret/myapp")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
