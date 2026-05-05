package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

// TrimResult holds the result of a trim operation for a single key.
type TrimResult struct {
	Key      string
	OldValue string
	NewValue string
	Changed  bool
}

// TrimSecrets reads all secrets at the given path, trims leading/trailing
// whitespace from all string values, and writes back the cleaned data.
// If dryRun is true, no writes are performed.
func TrimSecrets(ctx context.Context, client *api.Client, path string, dryRun bool) ([]TrimResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets at %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path %q", path)
	}

	raw := secret.Data
	// KV v2 wraps data under a "data" key.
	if inner, ok := raw["data"]; ok {
		if m, ok := inner.(map[string]interface{}); ok {
			raw = m
		}
	}

	results := make([]TrimResult, 0, len(raw))
	updated := make(map[string]interface{}, len(raw))

	for k, v := range raw {
		str, ok := v.(string)
		if !ok {
			updated[k] = v
			continue
		}
		trimmed := strings.TrimSpace(str)
		results = append(results, TrimResult{
			Key:      k,
			OldValue: str,
			NewValue: trimmed,
			Changed:  str != trimmed,
		})
		updated[k] = trimmed
	}

	if dryRun {
		return results, nil
	}

	_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{"data": updated})
	if err != nil {
		return nil, fmt.Errorf("failed to write trimmed secrets to %q: %w", path, err)
	}

	return results, nil
}
