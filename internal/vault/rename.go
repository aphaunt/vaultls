package vault

import (
	"context"
	"fmt"
)

// RenameSecrets copies secrets from srcPath to dstPath and optionally deletes the source.
// If overwrite is false and the destination already exists, an error is returned.
// The operation is not atomic: if writing to dstPath succeeds but deleting srcPath fails,
// the secrets will exist at both paths and the error will indicate the partial state.
func RenameSecrets(ctx context.Context, client *Client, srcPath, dstPath string, overwrite bool) error {
	if srcPath == dstPath {
		return fmt.Errorf("source and destination paths must differ: %q", srcPath)
	}

	// Read source secrets
	secrets, err := client.GetSecrets(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("reading source path %q: %w", srcPath, err)
	}

	if len(secrets) == 0 {
		return fmt.Errorf("source path %q is empty or does not exist", srcPath)
	}

	if !overwrite {
		// Check if destination already exists
		existing, err := client.GetSecrets(ctx, dstPath)
		if err == nil && len(existing) > 0 {
			return fmt.Errorf("destination path %q already exists; use --overwrite to replace", dstPath)
		}
	}

	// Write to destination
	if err := client.WriteSecrets(ctx, dstPath, secrets); err != nil {
		return fmt.Errorf("writing destination path %q: %w", dstPath, err)
	}

	// Delete source
	if err := client.DeleteSecrets(ctx, srcPath); err != nil {
		return fmt.Errorf("deleting source path %q after rename (secrets were written to %q): %w", srcPath, dstPath, err)
	}

	return nil
}
