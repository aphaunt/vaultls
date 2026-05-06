package vault

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/vault/api"
)

// FlattenSecrets reads all secrets at the given KV path and returns a flat
// map where nested keys are joined with the provided separator (e.g. "__").
// If dryRun is true the flattened result is returned without writing it back.
func FlattenSecrets(ctx context.Context, client *api.Client, path, separator string, dryRun bool) (map[string]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if separator == "" {
		separator = "__"
	}

	mount, sub := extractFlattenMount(path)
	dataPath := mount + "/data/" + sub

	secret, err := client.Logical().ReadWithContext(ctx, dataPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dataPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at %s", path)
	}

	raw, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at %s", path)
	}

	flat := make(map[string]string)
	flattenMap("", raw, separator, flat)

	if dryRun {
		return flat, nil
	}

	// Write the flattened map back, preserving existing metadata.
	payload := map[string]interface{}{"data": toStringInterface(flat)}
	_, err = client.Logical().WriteWithContext(ctx, dataPath, payload)
	if err != nil {
		return nil, fmt.Errorf("write %s: %w", dataPath, err)
	}

	return flat, nil
}

// FlatKeys returns a sorted slice of the flattened keys for display purposes.
func FlatKeys(flat map[string]string) []string {
	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func flattenMap(prefix string, m map[string]interface{}, sep string, out map[string]string) {
	for k, v := range m {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + sep + k
		}
		switch val := v.(type) {
		case map[string]interface{}:
			flattenMap(fullKey, val, sep, out)
		default:
			out[fullKey] = fmt.Sprintf("%v", val)
		}
	}
}

func toStringInterface(m map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func extractFlattenMount(path string) (string, string) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
