package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockMaskServer(data map[string]interface{}, written *map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodRead {
			resp := map[string]interface{}{"data": map[string]interface{}{"data": data}}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if written != nil {
				if d, ok := body["data"].(map[string]interface{}); ok {
					*written = d
				}
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func newMaskClient(server *httptest.Server) *vaultapi.Client {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = server.URL
	c, _ := vaultapi.NewClient(cfg)
	c.SetToken("test-token")
	return c
}

func TestMaskSecrets_EmptyPath(t *testing.T) {
	server := newMockMaskServer(map[string]interface{}{}, nil)
	defer server.Close()
	client := newMaskClient(server)

	_, err := MaskSecrets(context.Background(), client, "", []string{`password`}, "***", false)
	if err == nil || err.Error() != "path must not be empty" {
		t.Fatalf("expected empty path error, got %v", err)
	}
}

func TestMaskSecrets_NoPatterns(t *testing.T) {
	server := newMockMaskServer(map[string]interface{}{}, nil)
	defer server.Close()
	client := newMaskClient(server)

	_, err := MaskSecrets(context.Background(), client, "secret/myapp", nil, "***", false)
	if err == nil {
		t.Fatal("expected error for empty patterns")
	}
}

func TestMaskSecrets_InvalidPattern(t *testing.T) {
	server := newMockMaskServer(map[string]interface{}{"key": "value"}, nil)
	defer server.Close()
	client := newMaskClient(server)

	_, err := MaskSecrets(context.Background(), client, "secret/myapp", []string{`[invalid`}, "***", false)
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestMaskSecrets_DryRunDoesNotWrite(t *testing.T) {
	data := map[string]interface{}{"api_key": "supersecret123", "username": "admin"}
	var written map[string]interface{}
	server := newMockMaskServer(data, &written)
	defer server.Close()
	client := newMaskClient(server)

	count, err := MaskSecrets(context.Background(), client, "secret/myapp", []string{`supersecret\w+`}, "***", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 masked value, got %d", count)
	}
	if written != nil {
		t.Fatal("expected no write in dry-run mode")
	}
}

func TestMaskSecrets_WritesOnMatch(t *testing.T) {
	data := map[string]interface{}{"password": "hunter2", "user": "alice"}
	var written map[string]interface{}
	server := newMockMaskServer(data, &written)
	defer server.Close()
	client := newMaskClient(server)

	count, err := MaskSecrets(context.Background(), client, "secret/myapp", []string{`hunter\d`}, "[REDACTED]", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 masked value, got %d", count)
	}
}
