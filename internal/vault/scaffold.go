package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

// ScaffoldOptions configures scaffold behaviour.
type ScaffoldOptions struct {
	Path      string
	Keys      []string
	Defaults  map[string]string
	DryRun    bool
	Overwrite bool
}

// ScaffoldSecrets writes a skeleton secret at the given KV v2 path,
// populating each key with a default value (empty string if not provided).
func ScaffoldSecrets(ctx context.Context, client *api.Client, opts ScaffoldOptions) (map[string]string, error) {
	if strings.TrimSpace(opts.Path) == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if len(opts.Keys) == 0 {
		return nil, fmt.Errorf("at least one key must be specified")
	}

	mount, sub := extractScaffoldMount(opts.Path)
	dataPath := mount + "/data/" + sub

	// Check existence unless overwrite is set.
	if !opts.Overwrite {
		existing, err := client.Logical().ReadWithContext(ctx, dataPath)
		if err != nil {
			return nil, fmt.Errorf("scaffold existence check failed: %w", err)
		}
		if existing != nil {
			return nil, fmt.Errorf("secret already exists at %q; use --overwrite to replace", opts.Path)
		}
	}

	data := make(map[string]interface{})
	result := make(map[string]string)
	for _, k := range opts.Keys {
		v := ""
		if dv, ok := opts.Defaults[k]; ok {
			v = dv
		}
		data[k] = v
		result[k] = v
	}

	if opts.DryRun {
		return result, nil
	}

	_, err := client.Logical().WriteWithContext(ctx, dataPath, map[string]interface{}{"data": data})
	if err != nil {
		return nil, fmt.Errorf("scaffold write failed: %w", err)
	}
	return result, nil
}

func extractScaffoldMount(path string) (string, string) {
	parts := strings.SplitN(strings.Trim(path, "/"), "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
