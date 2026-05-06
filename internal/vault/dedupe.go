package vault

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/vault/api"
)

// DedupeResult holds information about a deduplicated secret.
type DedupeResult struct {
	Key        string
	Duplicates []string
}

// DedupeSecrets scans secrets at the given path, finds keys with identical
// values, and optionally removes duplicates keeping only the first occurrence
// (sorted alphabetically).
func DedupeSecrets(ctx context.Context, client *api.Client, path string, dryRun bool) ([]DedupeResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	dataPath := "secret/data/" + path
	secret, err := client.Logical().ReadWithContext(ctx, dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets at %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at %q", path)
	}

	rawData, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at %q", path)
	}

	// Build value -> keys mapping
	valueMap := make(map[string][]string)
	for k, v := range rawData {
		val := fmt.Sprintf("%v", v)
		valueMap[val] = append(valueMap[val], k)
	}

	var results []DedupeResult
	clean := make(map[string]interface{})

	for _, keys := range valueMap {
		sort.Strings(keys)
		if len(keys) > 1 {
			results = append(results, DedupeResult{
				Key:        keys[0],
				Duplicates: keys[1:],
			})
		}
		clean[keys[0]] = rawData[keys[0]]
	}

	// Add non-duplicate keys
	for k, v := range rawData {
		if _, exists := clean[k]; !exists {
			clean[k] = v
		}
	}

	if len(results) == 0 || dryRun {
		return results, nil
	}

	_, err = client.Logical().WriteWithContext(ctx, dataPath, map[string]interface{}{
		"data": clean,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to write deduplicated secrets to %q: %w", path, err)
	}

	return results, nil
}
