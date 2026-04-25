package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// TouchResult holds the result of a touch operation on a single secret key.
type TouchResult struct {
	Key     string
	Success bool
	Error   error
}

// TouchSecrets re-writes secrets at the given path to refresh their metadata
// (e.g. updated_time) without changing values. Useful for triggering
// lease renewals or audit events.
func TouchSecrets(ctx context.Context, client *api.Client, path string) ([]TouchResult, error) {
	// Read existing secrets
	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading path %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path %q", path)
	}

	data, ok := secret.Data["data"]
	if !ok {
		// Fallback for non-KV v2
		data = secret.Data
	}

	kvData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at path %q", path)
	}

	// Re-write the same data back
	_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{
		"data": kvData,
	})
	if err != nil {
		return nil, fmt.Errorf("writing path %q: %w", path, err)
	}

	results := make([]TouchResult, 0, len(kvData))
	for k := range kvData {
		results = append(results, TouchResult{
			Key:     k,
			Success: true,
		})
	}
	return results, nil
}
