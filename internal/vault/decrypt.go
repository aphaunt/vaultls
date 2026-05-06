package vault

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"

	"github.com/hashicorp/vault/api"
)

// DecryptSecrets reads secrets at path, decrypts values matching the given
// patterns using Vault's Transit engine, and writes them back (unless dry-run).
func DecryptSecrets(ctx context.Context, client *api.Client, path, transitKey string, patterns []string, dryRun bool) (map[string]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if transitKey == "" {
		return nil, fmt.Errorf("transit key must not be empty")
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("at least one pattern must be provided")
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", p, err)
		}
		compiled = append(compiled, re)
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading secret: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at %q", path)
	}

	raw, _ := secret.Data["data"].(map[string]interface{})
	if raw == nil {
		raw = secret.Data
	}

	updated := make(map[string]string)
	for k, v := range raw {
		str, ok := v.(string)
		if !ok {
			continue
		}
		if !matchesPatterns(compiled, k) {
			updated[k] = str
			continue
		}
		decrypted, err := decryptViaTransit(ctx, client, transitKey, str)
		if err != nil {
			return nil, fmt.Errorf("decrypting key %q: %w", k, err)
		}
		updated[k] = decrypted
	}

	if !dryRun {
		data := make(map[string]interface{}, len(updated))
		for k, v := range updated {
			data[k] = v
		}
		_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{"data": data})
		if err != nil {
			return nil, fmt.Errorf("writing decrypted secret: %w", err)
		}
	}

	return updated, nil
}

func decryptViaTransit(ctx context.Context, client *api.Client, key, ciphertext string) (string, error) {
	res, err := client.Logical().WriteWithContext(ctx, fmt.Sprintf("transit/decrypt/%s", key), map[string]interface{}{
		"ciphertext": ciphertext,
	})
	if err != nil {
		return "", err
	}
	if res == nil || res.Data == nil {
		return "", fmt.Errorf("empty response from transit decrypt")
	}
	plainB64, ok := res.Data["plaintext"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected plaintext type")
	}
	bytes, err := base64.StdEncoding.DecodeString(plainB64)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	return string(bytes), nil
}
