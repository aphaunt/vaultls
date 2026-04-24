package vault

import (
	"context"
	"fmt"
)

// CopyResult holds the outcome of a copy operation.
type CopyResult struct {
	Source      string
	Destination string
	KeysCopied  int
}

// CopySecrets reads all key-value pairs from srcPath and writes them to dstPath.
// If overwrite is false, existing keys at the destination are preserved.
func CopySecrets(ctx context.Context, client *Client, srcPath, dstPath string, overwrite bool) (*CopyResult, error) {
	srcData, err := client.GetSecret(ctx, srcPath)
	if err != nil {
		return nil, fmt.Errorf("reading source %q: %w", srcPath, err)
	}

	var dstData map[string]interface{}
	if !overwrite {
		existing, err := client.GetSecret(ctx, dstPath)
		if err == nil {
			dstData = existing
		} else {
			dstData = make(map[string]interface{})
		}
	} else {
		dstData = make(map[string]interface{})
	}

	copied := 0
	for k, v := range srcData {
		if _, exists := dstData[k]; exists && !overwrite {
			continue
		}
		dstData[k] = v
		copied++
	}

	if err := client.WriteSecret(ctx, dstPath, dstData); err != nil {
		return nil, fmt.Errorf("writing destination %q: %w", dstPath, err)
	}

	return &CopyResult{
		Source:      srcPath,
		Destination: dstPath,
		KeysCopied:  copied,
	}, nil
}
