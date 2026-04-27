package vault

import (
	"context"
	"fmt"
	"strings"
)

// TagEntry represents a secret path with its associated tags.
type TagEntry struct {
	Path string
	Tags map[string]string
}

// TagSecrets reads the secret at path, merges the provided tags into its
// metadata custom_metadata field, and writes it back. When overwrite is false
// existing tag keys are preserved.
func TagSecrets(ctx context.Context, client *Client, path string, tags map[string]string, overwrite bool) error {
	if path == "" {
		return fmt.Errorf("path must not be empty")
	}
	if len(tags) == 0 {
		return fmt.Errorf("at least one tag must be provided")
	}

	mount, subPath := splitMount(path)
	metaPath := fmt.Sprintf("%s/metadata/%s", mount, subPath)

	existing, err := client.vault.KVv2(mount).GetMetadata(ctx, subPath)
	if err != nil && !strings.Contains(err.Error(), "secret not found") {
		return fmt.Errorf("read metadata %s: %w", path, err)
	}

	current := map[string]string{}
	if existing != nil {
		for k, v := range existing.CustomMetadata {
			current[k] = fmt.Sprintf("%v", v)
		}
	}

	for k, v := range tags {
		if _, exists := current[k]; exists && !overwrite {
			continue
		}
		current[k] = v
	}

	_ = metaPath
	err = client.vault.KVv2(mount).PutMetadata(ctx, subPath, current)
	if err != nil {
		return fmt.Errorf("write metadata %s: %w", path, err)
	}
	return nil
}

// ListTags returns the custom metadata tags for the secret at path.
func ListTags(ctx context.Context, client *Client, path string) (map[string]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	mount, subPath := splitMount(path)
	meta, err := client.vault.KVv2(mount).GetMetadata(ctx, subPath)
	if err != nil {
		return nil, fmt.Errorf("read metadata %s: %w", path, err)
	}
	tags := map[string]string{}
	for k, v := range meta.CustomMetadata {
		tags[k] = fmt.Sprintf("%v", v)
	}
	return tags, nil
}
