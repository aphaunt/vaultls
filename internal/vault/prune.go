package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// PruneResult holds the result of a prune operation.
type PruneResult struct {
	Deleted []string
	Skipped []string
	DryRun  bool
}

// PruneSecrets removes secrets at the given path whose values are empty or match
// the provided patterns. If dryRun is true, no deletions are performed.
func PruneSecrets(ctx context.Context, client *vaultapi.Client, path string, patterns []string, dryRun bool) (*PruneResult, error) {
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

	data, ok := secret.Data["data"]
	if !ok {
		data = secret.Data
	}

	kv, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at path %q", path)
	}

	result := &PruneResult{DryRun: dryRun}
	updated := make(map[string]interface{})

	for k, v := range kv {
		val, _ := v.(string)
		if shouldPrune(k, val, patterns) {
			result.Deleted = append(result.Deleted, k)
		} else {
			result.Skipped = append(result.Skipped, k)
			updated[k] = v
		}
	}

	if !dryRun && len(result.Deleted) > 0 {
		writePath := path
		payload := map[string]interface{}{"data": updated}
		if _, err := client.Logical().WriteWithContext(ctx, writePath, payload); err != nil {
			return nil, fmt.Errorf("failed to write pruned secrets to %q: %w", path, err)
		}
	}

	return result, nil
}

func shouldPrune(key, value string, patterns []string) bool {
	if strings.TrimSpace(value) == "" {
		return true
	}
	for _, p := range patterns {
		if strings.Contains(key, p) || strings.Contains(value, p) {
			return true
		}
	}
	return false
}
