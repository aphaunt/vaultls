package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// SquashOptions controls how secrets are squashed into a single path.
type SquashOptions struct {
	Paths     []string
	Dest      string
	Prefix    bool
	DryRun    bool
	Overwrite bool
}

// SquashSecrets merges multiple KV secret paths into a single destination path.
// Keys from later paths overwrite earlier ones unless Overwrite is false.
func SquashSecrets(ctx context.Context, client *vaultapi.Client, opts SquashOptions) (map[string]string, error) {
	if len(opts.Paths) == 0 {
		return nil, fmt.Errorf("at least one source path is required")
	}
	if opts.Dest == "" {
		return nil, fmt.Errorf("destination path is required")
	}
	for _, p := range opts.Paths {
		if p == opts.Dest {
			return nil, fmt.Errorf("source path %q must not equal destination", p)
		}
	}

	merged := make(map[string]interface{})

	for _, src := range opts.Paths {
		secret, err := client.Logical().ReadWithContext(ctx, src)
		if err != nil {
			return nil, fmt.Errorf("reading %q: %w", src, err)
		}
		if secret == nil || secret.Data == nil {
			continue
		}
		data, ok := secret.Data["data"].(map[string]interface{})
		if !ok {
			data = secret.Data
		}
		for k, v := range data {
			key := k
			if opts.Prefix {
				segments := strings.Split(strings.Trim(src, "/"), "/")
				key = segments[len(segments)-1] + "_" + k
			}
			if _, exists := merged[key]; !exists || opts.Overwrite {
				merged[key] = v
			}
		}
	}

	result := make(map[string]string, len(merged))
	for k, v := range merged {
		result[k] = fmt.Sprintf("%v", v)
	}

	if opts.DryRun {
		return result, nil
	}

	payload := map[string]interface{}{"data": merged}
	_, err := client.Logical().WriteWithContext(ctx, opts.Dest, payload)
	if err != nil {
		return nil, fmt.Errorf("writing squashed secrets to %q: %w", opts.Dest, err)
	}

	return result, nil
}
