package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// PatchResult holds the outcome of a patch operation on a single key.
type PatchResult struct {
	Key     string
	OldVal  string
	NewVal  string
	Skipped bool
	Reason  string
}

// PatchSecrets applies selective key-value updates to a Vault KV path.
// Only the keys present in updates are modified; all other keys are preserved.
// If dryRun is true, no writes are performed.
func PatchSecrets(ctx context.Context, client *api.Client, path string, updates map[string]string, dryRun bool) ([]PatchResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if len(updates) == 0 {
		return nil, fmt.Errorf("updates must not be empty")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	existing := map[string]interface{}{}
	if secret != nil && secret.Data != nil {
		if kv, ok := secret.Data["data"]; ok {
			if m, ok := kv.(map[string]interface{}); ok {
				existing = m
			}
		} else {
			existing = secret.Data
		}
	}

	var results []PatchResult
	merged := make(map[string]interface{}, len(existing))
	for k, v := range existing {
		merged[k] = v
	}

	for key, newVal := range updates {
		oldVal := ""
		if v, ok := existing[key]; ok {
			oldVal = fmt.Sprintf("%v", v)
		}
		results = append(results, PatchResult{
			Key:    key,
			OldVal: oldVal,
			NewVal: newVal,
		})
		merged[key] = newVal
	}

	if dryRun {
		for i := range results {
			results[i].Skipped = true
			results[i].Reason = "dry-run"
		}
		return results, nil
	}

	_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{"data": merged})
	if err != nil {
		return nil, fmt.Errorf("write %s: %w", path, err)
	}

	return results, nil
}
