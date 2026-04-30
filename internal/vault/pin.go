package vault

import (
	"context"
	"fmt"
	"strconv"

	vaultapi "github.com/hashicorp/vault/api"
)

const pinMetaKey = "_pinned_version"

// PinSecret pins a secret path to a specific version by storing metadata.
func PinSecret(ctx context.Context, client *vaultapi.Client, path string, version int) error {
	if path == "" {
		return fmt.Errorf("path must not be empty")
	}
	if version < 1 {
		return fmt.Errorf("version must be >= 1")
	}

	existing, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read secret at %s: %w", path, err)
	}
	if existing == nil {
		return fmt.Errorf("no secret found at path: %s", path)
	}

	data, ok := existing.Data["data"].(map[string]interface{})
	if !ok {
		data = make(map[string]interface{})
	}
	data[pinMetaKey] = strconv.Itoa(version)

	_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{"data": data})
	if err != nil {
		return fmt.Errorf("failed to pin secret at %s: %w", path, err)
	}
	return nil
}

// GetPin returns the pinned version for a secret, or 0 if not pinned.
func GetPin(ctx context.Context, client *vaultapi.Client, path string) (int, error) {
	if path == "" {
		return 0, fmt.Errorf("path must not be empty")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return 0, fmt.Errorf("failed to read secret at %s: %w", path, err)
	}
	if secret == nil {
		return 0, nil
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return 0, nil
	}

	raw, ok := data[pinMetaKey]
	if !ok {
		return 0, nil
	}

	v, err := strconv.Atoi(fmt.Sprintf("%v", raw))
	if err != nil {
		return 0, fmt.Errorf("invalid pinned version value: %v", raw)
	}
	return v, nil
}

// UnpinSecret removes the pinned version marker from a secret.
func UnpinSecret(ctx context.Context, client *vaultapi.Client, path string) error {
	if path == "" {
		return fmt.Errorf("path must not be empty")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read secret at %s: %w", path, err)
	}
	if secret == nil {
		return fmt.Errorf("no secret found at path: %s", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil
	}
	delete(data, pinMetaKey)

	_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{"data": data})
	if err != nil {
		return fmt.Errorf("failed to unpin secret at %s: %w", path, err)
	}
	return nil
}
