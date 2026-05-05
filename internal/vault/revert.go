package vault

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// RevertSecrets restores a KV v2 secret at path to a specific prior version.
// If dryRun is true, it reads and validates the target version without writing.
func RevertSecrets(ctx context.Context, client *vaultapi.Client, path string, version int, dryRun bool) (map[string]interface{}, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client is nil")
	}
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if version < 1 {
		return nil, fmt.Errorf("version must be >= 1, got %d", version)
	}

	mount, sub := splitRevertMount(path)
	readPath := fmt.Sprintf("%s/data/%s", mount, sub)

	// Read the target version
	secret, err := client.Logical().ReadWithDataWithContext(ctx, readPath, map[string][]string{
		"version": {strconv.Itoa(version)},
	})
	if err != nil {
		return nil, fmt.Errorf("reading version %d at %s: %w", version, path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("version %d not found at %s", version, path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at %s version %d", path, version)
	}

	if dryRun {
		return data, nil
	}

	// Write the historical data as a new version
	writePath := fmt.Sprintf("%s/data/%s", mount, sub)
	_, err = client.Logical().WriteWithContext(ctx, writePath, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		return nil, fmt.Errorf("writing reverted data to %s: %w", path, err)
	}

	return data, nil
}

// splitRevertMount splits a KV v2 path into mount and sub-path.
func splitRevertMount(path string) (string, string) {
	parts := strings.SplitN(strings.Trim(path, "/"), "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
