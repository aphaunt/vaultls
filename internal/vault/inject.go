package vault

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
)

// InjectSecrets reads secrets from the given Vault path and injects them
// into the current process environment. If dryRun is true, it prints
// the variables that would be set without modifying the environment.
func InjectSecrets(ctx context.Context, client *api.Client, path string, dryRun bool) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets at %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path %q", path)
	}

	data, ok := secret.Data["data"]
	if !ok {
		data = secret.Data
	}

	kv, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at path %q", path)
	}

	var injected []string
	for k, v := range kv {
		envKey := strings.ToUpper(k)
		envVal := fmt.Sprintf("%v", v)
		injected = append(injected, fmt.Sprintf("%s=%s", envKey, envVal))
		if !dryRun {
			if err := os.Setenv(envKey, envVal); err != nil {
				return nil, fmt.Errorf("failed to set env var %q: %w", envKey, err)
			}
		}
	}

	return injected, nil
}
