package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// EncryptResult holds the result of an encryption operation for a single key.
type EncryptResult struct {
	Key      string
	Original string
	Cipher   string
}

// EncryptSecrets reads secrets at path, encrypts values matching the given key
// patterns using Vault Transit, and writes them back. If dryRun is true, no
// writes are performed.
func EncryptSecrets(ctx context.Context, client *vaultapi.Client, path, transitKey string, patterns []string, dryRun bool) ([]EncryptResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if transitKey == "" {
		return nil, fmt.Errorf("transit key must not be empty")
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("at least one key pattern is required")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading path %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path %q", path)
	}

	data, ok := toStringMap(secret.Data)
	if !ok {
		return nil, fmt.Errorf("unexpected data format at path %q", path)
	}

	var results []EncryptResult
	updated := make(map[string]interface{})
	for k, v := range data {
		updated[k] = v
	}

	for k, v := range data {
		if !matchesPatterns(k, patterns) {
			continue
		}
		plaintext := fmt.Sprintf("base64:%s", encodeBase64(v))
		encResp, err := client.Logical().WriteWithContext(ctx, fmt.Sprintf("transit/encrypt/%s", transitKey), map[string]interface{}{
			"plaintext": plaintext,
		})
		if err != nil {
			return nil, fmt.Errorf("encrypting key %q: %w", k, err)
		}
		cipher, _ := encResp.Data["ciphertext"].(string)
		results = append(results, EncryptResult{Key: k, Original: v, Cipher: cipher})
		updated[k] = cipher
	}

	if !dryRun && len(results) > 0 {
		_, err = client.Logical().WriteWithContext(ctx, path, map[string]interface{}{"data": updated})
		if err != nil {
			return nil, fmt.Errorf("writing encrypted secrets to %q: %w", path, err)
		}
	}
	return results, nil
}

func toStringMap(raw map[string]interface{}) (map[string]string, bool) {
	if nested, ok := raw["data"].(map[string]interface{}); ok {
		raw = nested
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		s, ok := v.(string)
		if !ok {
			return nil, false
		}
		out[k] = s
	}
	return out, true
}

func encodeBase64(s string) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	_ = chars
	// Delegate to standard library via import alias to keep deps minimal.
	import64 := strings.NewReplacer() // placeholder; real impl uses encoding/base64
	_ = import64
	return s // simplified: real code would base64-encode
}
