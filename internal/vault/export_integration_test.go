//go:build integration
// +build integration

package vault

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newMockExportServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/secret/data/myapp") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"data":{"USERNAME":"admin","PASSWORD":"s3cr3t"}}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestExportIntegration_JSONFormat(t *testing.T) {
	srv := newMockExportServer(t)
	defer srv.Close()

	t.Setenv("VAULT_ADDR", srv.URL)
	t.Setenv("VAULT_TOKEN", "test-token")

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	secrets, err := client.GetSecrets("secret/data/myapp")
	if err != nil {
		t.Fatalf("GetSecrets error: %v", err)
	}

	var buf bytes.Buffer
	if err := ExportSecrets(&buf, secrets, FormatJSON); err != nil {
		t.Fatalf("ExportSecrets error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "USERNAME") {
		t.Errorf("expected USERNAME in JSON output, got: %s", out)
	}
}
