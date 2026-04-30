package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// GCResult holds the result of a garbage collection run.
type GCResult struct {
	Path    string
	Deleted bool
	DryRun  bool
}

// GCSecrets removes secrets at paths whose keys all match a set of patterns
// (e.g. keys that look like temporary or test secrets). If dryRun is true,
// no writes are performed.
func GCSecrets(ctx context.Context, client *vaultapi.Client, path string, patterns []string, dryRun bool) ([]GCResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("at least one pattern must be provided")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading path %q: %w", path, err)
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

	var results []GCResult
	for key := range kv {
		if matchesAny(key, patterns) {
			result := GCResult{Path: path + "/" + key, DryRun: dryRun}
			if !dryRun {
				delete(kv, key)
				result.Deleted = true
			}
			results = append(results, result)
		}
	}

	if !dryRun && len(results) > 0 {
		_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{"data": kv})
		if err != nil {
			return nil, fmt.Errorf("writing cleaned secrets to %q: %w", path, err)
		}
	}

	return results, nil
}

func matchesAny(key string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(key, p) {
			return true
		}
	}
	return false
}
