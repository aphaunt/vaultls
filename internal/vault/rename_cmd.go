package vault

import (
	"context"
	"fmt"
)

// RenameSecretsWithValidation wraps RenameSecrets with pre-flight checks
// and returns a summary of the operation.
type RenameSummary struct {
	Source      string
	Destination string
	KeysMoved   int
	Overwritten bool
}

func RenameSecretsWithValidation(ctx context.Context, client *Client, src, dst string, overwrite bool) (*RenameSummary, error) {
	if src == "" {
		return nil, fmt.Errorf("source path must not be empty")
	}
	if dst == "" {
		return nil, fmt.Errorf("destination path must not be empty")
	}
	if src == dst {
		return nil, fmt.Errorf("source and destination paths are identical: %q", src)
	}

	// List keys at source to count them
	keys, err := client.List(ctx, src)
	if err != nil {
		return nil, fmt.Errorf("failed to list source path %q: %w", src, err)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no secrets found at source path %q", src)
	}

	// Check if destination already has keys
	dstKeys, _ := client.List(ctx, dst)
	hasExisting := len(dstKeys) > 0
	if hasExisting && !overwrite {
		return nil, fmt.Errorf("destination %q already exists; use --overwrite to replace", dst)
	}

	if err := RenameSecrets(ctx, client, src, dst, overwrite); err != nil {
		return nil, err
	}

	return &RenameSummary{
		Source:      src,
		Destination: dst,
		KeysMoved:   len(keys),
		Overwritten: hasExisting && overwrite,
	}, nil
}
