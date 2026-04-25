package vault

import (
	"context"
	"fmt"
)

// CloneOptions configures the behaviour of CloneSecrets.
type CloneOptions struct {
	Overwrite bool
	DryRun    bool
}

// CloneResult summarises what happened during a clone operation.
type CloneResult struct {
	Copied  []string
	Skipped []string
}

// CloneSecrets duplicates all secrets from srcPath to dstPath.
// It respects Overwrite and DryRun options.
func CloneSecrets(ctx context.Context, client *Client, srcPath, dstPath string, opts CloneOptions) (*CloneResult, error) {
	if srcPath == dstPath {
		return nil, fmt.Errorf("source and destination paths must differ")
	}

	keys, err := client.List(ctx, srcPath)
	if err != nil {
		return nil, fmt.Errorf("listing source path %q: %w", srcPath, err)
	}

	result := &CloneResult{}

	for _, key := range keys {
		srcKey := srcPath + "/" + key
		dstKey := dstPath + "/" + key

		secret, err := client.Read(ctx, srcKey)
		if err != nil {
			return nil, fmt.Errorf("reading %q: %w", srcKey, err)
		}

		if !opts.Overwrite {
			existing, _ := client.Read(ctx, dstKey)
			if existing != nil {
				result.Skipped = append(result.Skipped, key)
				continue
			}
		}

		if !opts.DryRun {
			if err := client.Write(ctx, dstKey, secret); err != nil {
				return nil, fmt.Errorf("writing %q: %w", dstKey, err)
			}
		}

		result.Copied = append(result.Copied, key)
	}

	return result, nil
}
