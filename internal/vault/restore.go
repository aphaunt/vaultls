package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// RestoreSecrets reads secrets from a previously exported JSON snapshot and
// writes them back to the given destination path in Vault. If dryRun is true
// the function prints what would be written without making any changes.
func RestoreSecrets(ctx context.Context, client *vaultapi.Client, destPath, snapshotJSON string, overwrite, dryRun bool) ([]string, error) {
	if destPath == "" {
		return nil, fmt.Errorf("destination path must not be empty")
	}
	if snapshotJSON == "" {
		return nil, fmt.Errorf("snapshot JSON must not be empty")
	}

	secrets, err := snapshotFromJSON(snapshotJSON)
	if err != nil {
		return nil, fmt.Errorf("parse snapshot: %w", err)
	}

	var restored []string
	for key, val := range secrets {
		full := strings.TrimRight(destPath, "/") + "/" + key
		mount, sub := splitRestoreMount(full)
		dataPath := mount + "/data/" + sub

		if !overwrite {
			existing, err := client.Logical().ReadWithContext(ctx, dataPath)
			if err == nil && existing != nil && existing.Data != nil {
				continue
			}
		}

		if dryRun {
			restored = append(restored, full)
			continue
		}

		_, err := client.Logical().WriteWithContext(ctx, dataPath, map[string]interface{}{
			"data": map[string]interface{}{"value": val},
		})
		if err != nil {
			return nil, fmt.Errorf("write %s: %w", full, err)
		}
		restored = append(restored, full)
	}
	return restored, nil
}

// splitRestoreMount splits a KV v2 path into mount and sub-path.
func splitRestoreMount(path string) (string, string) {
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	if len(parts) < 2 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
