package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// SyncResult holds the outcome of a sync operation for a single key.
type SyncResult struct {
	Key     string
	Action  string // "created", "updated", "skipped"
}

// SyncSecrets synchronises secrets from src to dst path.
// If overwrite is false, existing keys at dst are skipped.
// If dryRun is true, no writes are performed.
func SyncSecrets(ctx context.Context, client *api.Client, src, dst string, overwrite, dryRun bool) ([]SyncResult, error) {
	if src == "" || dst == "" {
		return nil, fmt.Errorf("src and dst paths must not be empty")
	}
	if src == dst {
		return nil, fmt.Errorf("src and dst paths must differ")
	}

	srcData, err := readKVSecrets(ctx, client, src)
	if err != nil {
		return nil, fmt.Errorf("reading src %q: %w", src, err)
	}

	dstData, err := readKVSecrets(ctx, client, dst)
	if err != nil {
		// dst may not exist yet; treat as empty
		dstData = map[string]interface{}{}
	}

	var results []SyncResult

	for k, v := range srcData {
		if _, exists := dstData[k]; exists && !overwrite {
			results = append(results, SyncResult{Key: k, Action: "skipped"})
			continue
		}

		action := "created"
		if _, exists := dstData[k]; exists {
			action = "updated"
		}

		if !dryRun {
			dstData[k] = v
		}
		results = append(results, SyncResult{Key: k, Action: action})
	}

	if !dryRun && len(results) > 0 {
		path := "secret/data/" + dst
		_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{
			"data": dstData,
		})
		if err != nil {
			return nil, fmt.Errorf("writing dst %q: %w", dst, err)
		}
	}

	return results, nil
}
