package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// RollbackResult holds the outcome of a rollback operation.
type RollbackResult struct {
	Path    string
	Version int
	Success bool
	Error   error
}

// RollbackSecret rolls back a KV v2 secret at the given path to the specified version.
func RollbackSecret(ctx context.Context, client *api.Client, path string, version int) (*RollbackResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if version < 1 {
		return nil, fmt.Errorf("version must be >= 1, got %d", version)
	}

	mount, secretPath := splitMount(path)
	kvClient := client.KVv2(mount)

	// Read the target version
	secret, err := kvClient.GetVersion(ctx, secretPath, version)
	if err != nil {
		return &RollbackResult{Path: path, Version: version, Success: false, Error: err}, err
	}
	if secret == nil || secret.Data == nil {
		err := fmt.Errorf("version %d not found at path %q", version, path)
		return &RollbackResult{Path: path, Version: version, Success: false, Error: err}, err
	}

	// Write the old data as the new current version
	_, err = kvClient.Put(ctx, secretPath, secret.Data)
	if err != nil {
		return &RollbackResult{Path: path, Version: version, Success: false, Error: err}, err
	}

	return &RollbackResult{Path: path, Version: version, Success: true}, nil
}

// splitMount splits a full path like "secret/foo/bar" into mount "secret" and path "foo/bar".
func splitMount(fullPath string) (string, string) {
	for i, c := range fullPath {
		if c == '/' {
			return fullPath[:i], fullPath[i+1:]
		}
	}
	return fullPath, ""
}
