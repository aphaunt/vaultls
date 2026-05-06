package vault

import (
	"context"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// SecretExpiry holds expiry metadata for a secret key.
type SecretExpiry struct {
	Key       string
	ExpiresAt time.Time
	Expired   bool
	TTL       time.Duration
}

// ExpireSecrets reads secrets at path and returns expiry info based on
// a custom metadata key "expires_at" (RFC3339 format). The path should
// be in the form "<mount>/<secret-path>", e.g. "secret/myapp/db".
// Keys without a parseable "expires_at" value are silently skipped.
func ExpireSecrets(ctx context.Context, client *vaultapi.Client, path string, now time.Time) ([]SecretExpiry, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	mount, sub := extractExpireMount(path)
	metaPath := fmt.Sprintf("%s/metadata/%s", mount, sub)

	secret, err := client.Logical().ReadWithContext(ctx, metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata at %s: %w", metaPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no metadata found at %s", metaPath)
	}

	customMeta, ok := secret.Data["custom_metadata"].(map[string]interface{})
	if !ok {
		return []SecretExpiry{}, nil
	}

	var results []SecretExpiry
	for key, val := range customMeta {
		rawExpiry, ok := val.(string)
		if !ok {
			continue
		}
		expAt, err := time.Parse(time.RFC3339, rawExpiry)
		if err != nil {
			continue
		}
		ttl := expAt.Sub(now)
		results = append(results, SecretExpiry{
			Key:       key,
			ExpiresAt: expAt,
			Expired:   now.After(expAt),
			TTL:       ttl,
		})
	}
	return results, nil
}

// FilterExpired returns only the SecretExpiry entries that have already expired.
func FilterExpired(secrets []SecretExpiry) []SecretExpiry {
	var expired []SecretExpiry
	for _, s := range secrets {
		if s.Expired {
			expired = append(expired, s)
		}
	}
	return expired
}

func extractExpireMount(path string) (string, string) {
	for i, c := range path {
		if c == '/' && i > 0 {
			return path[:i], path[i+1:]
		}
	}
	return path, ""
}
