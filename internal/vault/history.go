package vault

import (
	"context"
	"fmt"
	"sort"
	"strconv"
)

// SecretVersion represents a single version of a secret.
type SecretVersion struct {
	Version  int
	Data     map[string]string
	Deleted  bool
}

// ListVersions returns all available versions for a KV v2 secret path.
func (c *Client) ListVersions(ctx context.Context, path string) ([]int, error) {
	secret, err := c.vault.KVv2(c.mount).GetVersionsAsList(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("listing versions for %q: %w", path, err)
	}

	versions := make([]int, 0, len(secret))
	for _, v := range secret {
		versions = append(versions, v.Version)
	}
	sort.Ints(versions)
	return versions, nil
}

// GetVersion retrieves a specific version of a KV v2 secret.
func (c *Client) GetVersion(ctx context.Context, path string, version int) (*SecretVersion, error) {
	secret, err := c.vault.KVv2(c.mount).GetVersion(ctx, path, version)
	if err != nil {
		return nil, fmt.Errorf("getting version %d of %q: %w", version, path, err)
	}

	data := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		data[k] = fmt.Sprintf("%v", v)
	}

	return &SecretVersion{
		Version: version,
		Data:    data,
		Deleted: secret.VersionMetadata != nil && !secret.VersionMetadata.DeletionTime.IsZero(),
	}, nil
}

// DiffVersions computes the diff between two versions of the same secret path.
func (c *Client) DiffVersions(ctx context.Context, path string, vA, vB int) ([]DiffResult, error) {
	svA, err := c.GetVersion(ctx, path, vA)
	if err != nil {
		return nil, fmt.Errorf("fetching version %s: %w", strconv.Itoa(vA), err)
	}
	svB, err := c.GetVersion(ctx, path, vB)
	if err != nil {
		return nil, fmt.Errorf("fetching version %s: %w", strconv.Itoa(vB), err)
	}
	return DiffSecrets(svA.Data, svB.Data), nil
}
