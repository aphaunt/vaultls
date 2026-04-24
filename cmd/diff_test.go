package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"vaultls/internal/vault"
)

func TestRenderDiff_NoDifferences(t *testing.T) {
	var buf bytes.Buffer
	result := vault.DiffResult{
		Changed: make(map[string][2]string),
	}
	vault.RenderDiff(&buf, result, "envA", "envB", false)
	assert.Contains(t, buf.String(), "No differences found.")
}

func TestRenderDiff_OnlyInA(t *testing.T) {
	var buf bytes.Buffer
	result := vault.DiffResult{
		OnlyInA: []string{"secret_key"},
		Changed: make(map[string][2]string),
	}
	vault.RenderDiff(&buf, result, "staging", "prod", false)
	assert.Contains(t, buf.String(), "[staging only]")
	assert.Contains(t, buf.String(), "secret_key")
}

func TestRenderDiff_OnlyInB(t *testing.T) {
	var buf bytes.Buffer
	result := vault.DiffResult{
		OnlyInB: []string{"new_key"},
		Changed: make(map[string][2]string),
	}
	vault.RenderDiff(&buf, result, "staging", "prod", false)
	assert.Contains(t, buf.String(), "[prod only]")
	assert.Contains(t, buf.String(), "new_key")
}

func TestRenderDiff_ChangedValues(t *testing.T) {
	var buf bytes.Buffer
	result := vault.DiffResult{
		InBoth:  []string{"db_pass"},
		Changed: map[string][2]string{"db_pass": {"old_val", "new_val"}},
	}
	vault.RenderDiff(&buf, result, "staging", "prod", false)
	out := buf.String()
	assert.Contains(t, out, "db_pass")
	assert.Contains(t, out, "old_val")
	assert.Contains(t, out, "new_val")
}
