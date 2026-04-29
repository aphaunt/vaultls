package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

const protectMetaKey = "vaultls_protected"

// ProtectSecrets marks a secret path as protected, preventing accidental overwrites.
func ProtectSecrets(ctx context.Context, client *vaultapi.Client, path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path must not be empty")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read secret at %s: %w", path, err)
	}
	if secret == nil {
		return fmt.Errorf("no secret found at path: %s", path)
	}

	data := secret.Data
	if data == nil {
		data = map[string]interface{}{}
	}
	data[protectMetaKey] = "true"

	_, err = client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return fmt.Errorf("failed to protect secret at %s: %w", path, err)
	}
	return nil
}

// UnprotectSecrets removes the protection marker from a secret path.
func UnprotectSecrets(ctx context.Context, client *vaultapi.Client, path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path must not be empty")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read secret at %s: %w", path, err)
	}
	if secret == nil {
		return fmt.Errorf("no secret found at path: %s", path)
	}

	data := secret.Data
	if data != nil {
		delete(data, protectMetaKey)
	}

	_, err = client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return fmt.Errorf("failed to unprotect secret at %s: %w", path, err)
	}
	return nil
}

// IsProtected returns true if the secret at the given path has the protection marker.
func IsProtected(ctx context.Context, client *vaultapi.Client, path string) (bool, error) {
	if strings.TrimSpace(path) == "" {
		return false, fmt.Errorf("path must not be empty")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return false, fmt.Errorf("failed to read secret at %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return false, nil
	}

	val, ok := secret.Data[protectMetaKey]
	if !ok {
		return false, nil
	}
	return fmt.Sprintf("%v", val) == "true", nil
}
