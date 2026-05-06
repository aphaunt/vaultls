package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// TruncateOptions configures truncation behaviour.
type TruncateOptions struct {
	MaxLength int
	Suffix    string
	DryRun    bool
	Keys      []string // if non-empty, only truncate these keys
}

// TruncateSecrets reads all key/value pairs at path and truncates string
// values that exceed MaxLength characters. Truncated values are suffixed
// with Suffix (default "..."). When DryRun is true the vault is not mutated.
func TruncateSecrets(ctx context.Context, client *vaultapi.Client, path string, opts TruncateOptions) (map[string]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if opts.MaxLength <= 0 {
		return nil, fmt.Errorf("max-length must be greater than zero")
	}
	if opts.Suffix == "" {
		opts.Suffix = "..."
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

	keySet := make(map[string]struct{}, len(opts.Keys))
	for _, k := range opts.Keys {
		keySet[k] = struct{}{}
	}

	changed := make(map[string]string)
	updated := make(map[string]interface{})
	for k, v := range dataMap {
		updated[k] = v
	}

	for k, v := range dataMap {
		if len(opts.Keys) > 0 {
			if _, ok := keySet[k]; !ok {
				continue
			}
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		if len(s) > opts.MaxLength {
			truncated := s[:opts.MaxLength]
			if !strings.HasSuffix(truncated, opts.Suffix) {
				truncated += opts.Suffix
			}
			changed[k] = truncated
			updated[k] = truncated
		}
	}

	if len(changed) == 0 || opts.DryRun {
		return changed, nil
	}

	writePath := strings.Replace(path, "/metadata/", "/data/", 1)
	_, err = client.Logical().WriteWithContext(ctx, writePath, map[string]interface{}{"data": updated})
	if err != nil {
		return nil, fmt.Errorf("writing truncated secrets to %s: %w", writePath, err)
	}
	return changed, nil
}
