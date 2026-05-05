package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

// ResolveResult holds the resolved value and metadata for a secret key.
type ResolveResult struct {
	Path  string
	Key   string
	Value string
}

// ResolveSecrets resolves one or more "path#key" references and returns their
// current values from Vault. Each ref must be in the form "secret/path#key".
func ResolveSecrets(ctx context.Context, client *api.Client, refs []string) ([]ResolveResult, error) {
	if len(refs) == 0 {
		return nil, fmt.Errorf("no refs provided")
	}

	var results []ResolveResult
	for _, ref := range refs {
		parts := strings.SplitN(ref, "#", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid ref %q: must be in the form path#key", ref)
		}
		path, key := parts[0], parts[1]

		secret, err := client.Logical().ReadWithContext(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", path, err)
		}
		if secret == nil || secret.Data == nil {
			return nil, fmt.Errorf("path %q not found", path)
		}

		data := secret.Data
		// KV v2 wraps values under "data"
		if inner, ok := data["data"]; ok {
			if m, ok := inner.(map[string]interface{}); ok {
				data = m
			}
		}

		v, ok := data[key]
		if !ok {
			return nil, fmt.Errorf("key %q not found at path %q", key, path)
		}

		results = append(results, ResolveResult{
			Path:  path,
			Key:   key,
			Value: fmt.Sprintf("%v", v),
		})
	}
	return results, nil
}
