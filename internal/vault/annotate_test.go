package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockAnnotateServer(t *testing.T) (*httptest.Server, *vaultapi.Client) {
	t.Helper()
	meta := map[string]interface{}{}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/secret/metadata/myapp/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"custom_metadata": meta},
			})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if cm, ok := body["custom_metadata"].(map[string]interface{}); ok {
				for k, v := range cm {
					meta[k] = v
				}
			}
			w.WriteHeader(http.StatusNoContent)
		}
	})

	ts := httptest.NewServer(mux)
	cfg := vaultapi.DefaultConfig()
	cfg.Address = ts.URL
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	client.SetToken("test-token")
	return ts, client
}

func newAnnotateClient(t *testing.T) (*httptest.Server, *vaultapi.Client) {
	return newMockAnnotateServer(t)
}

func TestAnnotateSecrets_EmptyPath(t *testing.T) {
	_, client := newAnnotateClient(t)
	err := AnnotateSecrets(context.Background(), client, "", map[string]string{"owner": "team-a"}, false)
	if err == nil || err.Error() != "path must not be empty" {
		t.Fatalf("expected empty path error, got %v", err)
	}
}

func TestAnnotateSecrets_NoAnnotations(t *testing.T) {
	_, client := newAnnotateClient(t)
	err := AnnotateSecrets(context.Background(), client, "secret/myapp/config", map[string]string{}, false)
	if err == nil {
		t.Fatal("expected error for empty annotations")
	}
}

func TestAnnotateSecrets_DryRunDoesNotWrite(t *testing.T) {
	ts, client := newAnnotateClient(t)
	defer ts.Close()
	err := AnnotateSecrets(context.Background(), client, "secret/myapp/config", map[string]string{"env": "prod"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	anns, err := GetAnnotations(context.Background(), client, "secret/myapp/config")
	if err != nil {
		t.Fatalf("unexpected error reading annotations: %v", err)
	}
	if len(anns) != 0 {
		t.Fatalf("expected no annotations after dry run, got %v", anns)
	}
}

func TestAnnotateSecrets_WritesAndReads(t *testing.T) {
	ts, client := newAnnotateClient(t)
	defer ts.Close()
	err := AnnotateSecrets(context.Background(), client, "secret/myapp/config", map[string]string{"owner": "team-a", "env": "staging"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	anns, err := GetAnnotations(context.Background(), client, "secret/myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if anns["owner"] != "team-a" {
		t.Errorf("expected owner=team-a, got %q", anns["owner"])
	}
	if anns["env"] != "staging" {
		t.Errorf("expected env=staging, got %q", anns["env"])
	}
}

func TestGetAnnotations_EmptyPath(t *testing.T) {
	_, client := newAnnotateClient(t)
	_, err := GetAnnotations(context.Background(), client, "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}
