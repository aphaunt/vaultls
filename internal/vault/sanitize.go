package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// SanitizeResult holds the outcome of a sanitize operation for a single key.
type SanitizeResult struct {
	Key      string
	OldValue string
	NewValue string
	Changed  bool
}

// SanitizeSecrets trims whitespace from all string values at the given path.
// If dryRun is true, changes are computed but not written back to Vault.
func SanitizeSecrets(ctx context.Context, client *vaultapi.Client, path string, dryRun bool) ([]SanitizeResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret at %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path: %s", path)
	}

	data, ok := secret.Data["data"]
	if !ok {
		data = secret.Data
	}

	kv, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at path: %s", path)
	}

	results := make([]SanitizeResult, 0, len(kv))
	updated := make(map[string]interface{}, len(kv))

	for k, v := range kv {
		str, isStr := v.(string)
		if !isStr {
			updated[k] = v
			continue
		}
		trimmed := strings.TrimSpace(str)
		result := SanitizeResult{
			Key:      k,
			OldValue: str,
			NewValue: trimmed,
			Changed:  trimmed != str,
		}
		results = append(results, result)
		updated[k] = trimmed
	}

	if !dryRun {
		_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{"data": updated})
		if err != nil {
			return nil, fmt.Errorf("failed to write sanitized secrets to %s: %w", path, err)
		}
	}

	return results, nil
}
