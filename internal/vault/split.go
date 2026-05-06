package vault

import (
	"context"
	"fmt"
	"strings"

	hashivault "github.com/hashicorp/vault/api"
)

// SplitSecrets reads secrets from a source path and writes subsets of keys
// to multiple destination paths based on a provided key-to-destination mapping.
// Keys not present in the mapping are ignored unless a default destination is set.
func SplitSecrets(ctx context.Context, client *hashivault.Client, srcPath string, mapping map[string]string, defaultDest string, dryRun bool) (map[string][]string, error) {
	if srcPath == "" {
		return nil, fmt.Errorf("source path must not be empty")
	}
	if len(mapping) == 0 && defaultDest == "" {
		return nil, fmt.Errorf("mapping must not be empty when no default destination is set")
	}

	secret, err := client.Logical().ReadWithContext(ctx, srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read source path %q: %w", srcPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path %q", srcPath)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		data = secret.Data
	}

	// Group keys by destination
	destBuckets := map[string]map[string]interface{}{}
	for key, val := range data {
		dest, found := mapping[key]
		if !found {
			if defaultDest == "" {
				continue
			}
			dest = defaultDest
		}
		if _, exists := destBuckets[dest]; !exists {
			destBuckets[dest] = map[string]interface{}{}
		}
		destBuckets[dest][key] = val
	}

	result := map[string][]string{}
	for dest, bucket := range destBuckets {
		keys := make([]string, 0, len(bucket))
		for k := range bucket {
			keys = append(keys, k)
		}
		result[dest] = keys

		if dryRun {
			continue
		}

		writePath := toSplitDataPath(dest)
		_, err := client.Logical().WriteWithContext(ctx, writePath, map[string]interface{}{"data": bucket})
		if err != nil {
			return nil, fmt.Errorf("failed to write to destination %q: %w", dest, err)
		}
	}

	return result, nil
}

func toSplitDataPath(path string) string {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 {
		return parts[0] + "/data/" + parts[1]
	}
	return path
}
