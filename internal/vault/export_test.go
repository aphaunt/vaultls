package vault

import (
	"bytes"
	"strings"
	"testing"
)

func TestExportSecrets_JSON(t *testing.T) {
	secrets := map[string]string{"FOO": "bar", "BAZ": "qux"}
	var buf bytes.Buffer
	if err := ExportSecrets(&buf, secrets, FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"FOO"`) || !strings.Contains(out, `"bar"`) {
		t.Errorf("expected JSON output to contain keys and values, got: %s", out)
	}
}

func TestExportSecrets_YAML(t *testing.T) {
	secrets := map[string]string{"KEY": "value"}
	var buf bytes.Buffer
	if err := ExportSecrets(&buf, secrets, FormatYAML); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "KEY") || !strings.Contains(out, "value") {
		t.Errorf("expected YAML output to contain key and value, got: %s", out)
	}
}

func TestExportSecrets_Env(t *testing.T) {
	secrets := map[string]string{"DB_PASS": "secret123", "API_KEY": "abc"}
	var buf bytes.Buffer
	if err := ExportSecrets(&buf, secrets, FormatEnv); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "API_KEY=") || !strings.Contains(out, "DB_PASS=") {
		t.Errorf("expected env output to contain keys, got: %s", out)
	}
	// verify sorted order
	idxAPI := strings.Index(out, "API_KEY")
	idxDB := strings.Index(out, "DB_PASS")
	if idxAPI > idxDB {
		t.Errorf("expected API_KEY before DB_PASS in sorted env output")
	}
}

func TestExportSecrets_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	err := ExportSecrets(&buf, map[string]string{}, "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
}
