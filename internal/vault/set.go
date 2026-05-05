package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

// SetSecretResult holds the result of a set operation for a single key.
type SetSecretResult struct {
	Key     string
	Written bool
	Skipped bool
	Reason  string
}

// SetSecrets writes key=value pairs to the given Vault path.
// If dryRun is true, no writes are performed.
func SetSecrets(ctx context.Context, client *api.Client, path string, pairs map[string]string, dryRun bool) ([]SetSecretResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if len(pairs) == 0 {
		return nil, fmt.Errorf("at least one key=value pair is required")
	}

	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid path %q: expected <mount>/path", path)
	}
	mount := parts[0]
	subPath := parts[1]
	dataPath := mount + "/data/" + subPath

	// Read existing data to merge
	existing := map[string]interface{}{}
	secret, err := client.Logical().ReadWithContext(ctx, dataPath)
	if err == nil && secret != nil {
		if data, ok := secret.Data["data"].(map[string]interface{}); ok {
			existing = data
		}
	}

	var results []SetSecretResult
	for k, v := range pairs {
		result := SetSecretResult{Key: k}
		if dryRun {
			result.Skipped = true
			result.Reason = "dry-run"
		} else {
			existing[k] = v
			result.Written = true
		}
		results = append(results, result)
	}

	if !dryRun {
		_, err = client.Logical().WriteWithContext(ctx, dataPath, map[string]interface{}{
			"data": existing,
		})
		if err != nil {
			return nil, fmt.Errorf("writing secrets to %s: %w", path, err)
		}
	}

	return results, nil
}
