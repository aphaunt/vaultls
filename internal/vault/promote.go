package vault

import (
	"context"
	"fmt"
)

// PromoteResult holds the outcome of a promotion operation.
type PromoteResult struct {
	Copied    []string
	Skipped   []string
	Overwritten []string
}

// PromoteSecrets copies all secrets from srcPath to dstPath, optionally
// overwriting existing keys at the destination. It returns a PromoteResult
// summarising what happened.
func PromoteSecrets(ctx context.Context, client *Client, srcPath, dstPath string, overwrite bool) (*PromoteResult, error) {
	if srcPath == dstPath {
		return nil, fmt.Errorf("source and destination paths must differ: %q", srcPath)
	}

	srcKeys, err := client.List(ctx, srcPath)
	if err != nil {
		return nil, fmt.Errorf("listing source %q: %w", srcPath, err)
	}

	dstKeys, err := client.List(ctx, dstPath)
	if err != nil {
		dstKeys = []string{}
	}
	dstSet := toSet(dstKeys)

	srcSecrets, err := client.GetSecrets(ctx, srcPath)
	if err != nil {
		return nil, fmt.Errorf("reading source secrets %q: %w", srcPath, err)
	}

	result := &PromoteResult{}

	for _, key := range srcKeys {
		if _, exists := dstSet[key]; exists {
			if !overwrite {
				result.Skipped = append(result.Skipped, key)
				continue
			}
			result.Overwritten = append(result.Overwritten, key)
		} else {
			result.Copied = append(result.Copied, key)
		}
	}

	writeData := map[string]interface{}{}
	for _, key := range result.Copied {
		writeData[key] = srcSecrets[key]
	}
	for _, key := range result.Overwritten {
		writeData[key] = srcSecrets[key]
	}

	if len(writeData) > 0 {
		if err := client.WriteSecrets(ctx, dstPath, writeData); err != nil {
			return nil, fmt.Errorf("writing to destination %q: %w", dstPath, err)
		}
	}

	return result, nil
}
