package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockTagServer(t *testing.T) *httptest.Server {
	t.Helper()
	meta := map[string]interface{}{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/metadata/myapp/config":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"custom_metadata": meta},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/secret/metadata/myapp/config":
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if cm, ok := body["custom_metadata"].(map[string]interface{}); ok {
				for k, v := range cm {
					meta[k] = v
				}
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestTagSecrets_EmptyPath(t *testing.T) {
	client := &Client{}
	err := TagSecrets(context.Background(), client, "", map[string]string{"env": "prod"}, false)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestTagSecrets_NoTags(t *testing.T) {
	client := &Client{}
	err := TagSecrets(context.Background(), client, "secret/myapp/config", map[string]string{}, false)
	if err == nil {
		t.Fatal("expected error when no tags provided")
	}
}

func TestListTags_EmptyPath(t *testing.T) {
	client := &Client{}
	_, err := ListTags(context.Background(), client, "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestTagSecrets_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := &Client{}
	err := TagSecrets(ctx, client, "secret/myapp/config", map[string]string{"env": "prod"}, false)
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}
