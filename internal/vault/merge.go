package vault

import (
	"context"
	"fmt"
)

// MergeResult holds the outcome of a merge operation.
type MergeResult struct {
	Written  []string
	Skipped  []string
	Overwritten []string
}

// MergeSecrets merges secrets from src into dst.
// If overwrite is false, existing keys in dst are skipped.
// If dryRun is true, no writes are performed.
func MergeSecrets(ctx context.Context, client *Client, src, dst string, overwrite, dryRun bool) (*MergeResult, error) {
	if src == "" || dst == "" {
		return nil, fmt.Errorf("src and dst paths must not be empty")
	}
	if src == dst {
		return nil, fmt.Errorf("src and dst paths must differ")
	}

	srcSecrets, err := client.GetSecrets(ctx, src)
	if err != nil {
		return nil, fmt.Errorf("reading src %q: %w", src, err)
	}

	dstSecrets, err := client.GetSecrets(ctx, dst)
	if err != nil {
		return nil, fmt.Errorf("reading dst %q: %w", dst, err)
	}

	merged := make(map[string]interface{})
	for k, v := range dstSecrets {
		merged[k] = v
	}

	result := &MergeResult{}

	for k, v := range srcSecrets {
		if _, exists := dstSecrets[k]; exists {
			if !overwrite {
				result.Skipped = append(result.Skipped, k)
				continue
			}
			result.Overwritten = append(result.Overwritten, k)
		} else {
			result.Written = append(result.Written, k)
		}
		merged[k] = v
	}

	if !dryRun && (len(result.Written) > 0 || len(result.Overwritten) > 0) {
		if err := client.WriteSecrets(ctx, dst, merged); err != nil {
			return nil, fmt.Errorf("writing merged secrets to %q: %w", dst, err)
		}
	}

	return result, nil
}
