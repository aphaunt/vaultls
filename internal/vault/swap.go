package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// SwapSecrets atomically swaps the contents of two KV paths.
// If dryRun is true, no writes are performed.
func SwapSecrets(ctx context.Context, client *api.Client, pathA, pathB string, dryRun bool) error {
	if pathA == "" || pathB == "" {
		return fmt.Errorf("both paths must be non-empty")
	}
	if pathA == pathB {
		return fmt.Errorf("source and destination paths must differ")
	}

	readA, err := client.Logical().ReadWithContext(ctx, pathA)
	if err != nil {
		return fmt.Errorf("reading %s: %w", pathA, err)
	}
	if readA == nil {
		return fmt.Errorf("path not found: %s", pathA)
	}

	readB, err := client.Logical().ReadWithContext(ctx, pathB)
	if err != nil {
		return fmt.Errorf("reading %s: %w", pathB, err)
	}
	if readB == nil {
		return fmt.Errorf("path not found: %s", pathB)
	}

	dataA := extractData(readA.Data)
	dataB := extractData(readB.Data)

	if dryRun {
		return nil
	}

	if _, err := client.Logical().WriteWithContext(ctx, pathA, map[string]interface{}{"data": dataB}); err != nil {
		return fmt.Errorf("writing %s: %w", pathA, err)
	}
	if _, err := client.Logical().WriteWithContext(ctx, pathB, map[string]interface{}{"data": dataA}); err != nil {
		return fmt.Errorf("writing %s: %w", pathB, err)
	}

	return nil
}

func extractData(raw map[string]interface{}) map[string]interface{} {
	if d, ok := raw["data"]; ok {
		if m, ok := d.(map[string]interface{}); ok {
			return m
		}
	}
	return raw
}
