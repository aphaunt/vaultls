package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockTemplateServer(data map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		iface := make(map[string]interface{}, len(data))
		for k, v := range data {
			iface[k] = v
		}
		payload := map[string]interface{}{"data": map[string]interface{}{"data": iface}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
}

func newTemplateClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestRenderTemplate_Success(t *testing.T) {
	srv := newMockTemplateServer(map[string]string{"HOST": "localhost", "PORT": "5432"})
	defer srv.Close()

	client := newTemplateClient(t, srv.URL)
	res, err := RenderTemplate(context.Background(), client, "secret/data/db", "{{HOST}}:{{PORT}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Rendered != "localhost:5432" {
		t.Errorf("expected 'localhost:5432', got %q", res.Rendered)
	}
}

func TestRenderTemplate_EmptyPath(t *testing.T) {
	client := newTemplateClient(t, "http://127.0.0.1")
	_, err := RenderTemplate(context.Background(), client, "", "{{KEY}}")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestRenderTemplate_EmptyTemplate(t *testing.T) {
	client := newTemplateClient(t, "http://127.0.0.1")
	_, err := RenderTemplate(context.Background(), client, "secret/data/x", "")
	if err == nil {
		t.Fatal("expected error for empty template")
	}
}

func TestRenderTemplate_UnknownPlaceholder(t *testing.T) {
	srv := newMockTemplateServer(map[string]string{"FOO": "bar"})
	defer srv.Close()

	client := newTemplateClient(t, srv.URL)
	res, err := RenderTemplate(context.Background(), client, "secret/data/cfg", "value={{UNKNOWN}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Rendered != "value={{UNKNOWN}}" {
		t.Errorf("expected placeholder to remain, got %q", res.Rendered)
	}
}
