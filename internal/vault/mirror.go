package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// MirrorResult holds the outcome of a mirror operation.
type MirrorResult struct {
	Copied  []string
	Skipped []string
	Deleted []string
}

// MirrorSecrets makes dst an exact replica of src.
// Keys present in src but absent in dst are written.
// Keys present in both are overwritten.
// Keys present in dst but absent in src are deleted unless deleteOrphans is false.
func MirrorSecrets(ctx context.Context, client *api.Client, src, dst string, deleteOrphans, dryRun bool) (*MirrorResult, error) {
	if src == "" || dst == "" {
		return nil, fmt.Errorf("src and dst paths must not be empty")
	}
	if src == dst {
		return nil, fmt.Errorf("src and dst must be different paths")
	}

	srcData, err := readKV(ctx, client, src)
	if err != nil {
		return nil, fmt.Errorf("reading src %q: %w", src, err)
	}

	dstData, err := readKV(ctx, client, dst)
	if err != nil {
		dstData = map[string]interface{}{}
	}

	result := &MirrorResult{}

	for k, v := range srcData {
		if existing, ok := dstData[k]; ok && existing == v {
			result.Skipped = append(result.Skipped, k)
			continue
		}
		if !dryRun {
			if err := writeKV(ctx, client, dst, k, v); err != nil {
				return nil, fmt.Errorf("writing key %q to dst: %w", k, err)
			}
		}
		result.Copied = append(result.Copied, k)
	}

	if deleteOrphans {
		for k := range dstData {
			if _, ok := srcData[k]; !ok {
				if !dryRun {
					if err := deleteKV(ctx, client, dst, k); err != nil {
						return nil, fmt.Errorf("deleting orphan key %q from dst: %w", k, err)
					}
				}
				result.Deleted = append(result.Deleted, k)
			}
		}
	}

	return result, nil
}

func readKV(ctx context.Context, client *api.Client, path string) (map[string]interface{}, error) {
	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return map[string]interface{}{}, nil
	}
	if data, ok := secret.Data["data"].(map[string]interface{}); ok {
		return data, nil
	}
	return secret.Data, nil
}

func writeKV(ctx context.Context, client *api.Client, path, key string, value interface{}) error {
	existing, err := readKV(ctx, client, path)
	if err != nil {
		existing = map[string]interface{}{}
	}
	existing[key] = value
	_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{"data": existing})
	return err
}

func deleteKV(ctx context.Context, client *api.Client, path, key string) error {
	existing, err := readKV(ctx, client, path)
	if err != nil {
		return err
	}
	delete(existing, key)
	_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{"data": existing})
	return err
}
