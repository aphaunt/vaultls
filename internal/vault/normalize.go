package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

// NormalizeOptions controls how secrets are normalized.
type NormalizeOptions struct {
	DryRun    bool
	TrimSpace bool
	LowerKeys bool
	UpperKeys bool
}

// NormalizeSecrets reads secrets at path, applies normalization rules, and
// writes the result back unless DryRun is set. Returns the normalized map.
func NormalizeSecrets(ctx context.Context, client *api.Client, path string, opts NormalizeOptions) (map[string]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if opts.LowerKeys && opts.UpperKeys {
		return nil, fmt.Errorf("LowerKeys and UpperKeys are mutually exclusive")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at %s", path)
	}

	raw, ok := secret.Data["data"]
	if !ok {
		raw = secret.Data
	}
	dataMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at %s", path)
	}

	normalized := make(map[string]string, len(dataMap))
	for k, v := range dataMap {
		key := k
		val := fmt.Sprintf("%v", v)

		if opts.TrimSpace {
			key = strings.TrimSpace(key)
			val = strings.TrimSpace(val)
		}
		if opts.LowerKeys {
			key = strings.ToLower(key)
		}
		if opts.UpperKeys {
			key = strings.ToUpper(key)
		}
		normalized[key] = val
	}

	if !opts.DryRun {
		payload := make(map[string]interface{}, len(normalized))
		for k, v := range normalized {
			payload[k] = v
		}
		_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{"data": payload})
		if err != nil {
			return nil, fmt.Errorf("writing normalized secrets to %s: %w", path, err)
		}
	}

	return normalized, nil
}
