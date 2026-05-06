package vault

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/vault/api"
)

// DigestResult holds the computed digest for a secret path.
type DigestResult struct {
	Path   string
	Digest string
	Keys   int
}

// DigestSecrets computes a stable SHA-256 digest over all key-value pairs
// at the given KV path. The digest is deterministic regardless of key order.
func DigestSecrets(ctx context.Context, client *api.Client, path string) (*DigestResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	mount, sub := extractDigestMount(path)
	dataPath := mount + "/data/" + sub

	secret, err := client.Logical().ReadWithContext(ctx, dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret at %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path: %s", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at path: %s", path)
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		v := fmt.Sprintf("%v", data[k])
		fmt.Fprintf(h, "%s=%s\n", k, v)
	}

	digest := fmt.Sprintf("%x", h.Sum(nil))
	return &DigestResult{
		Path:   path,
		Digest: digest,
		Keys:   len(keys),
	}, nil
}

// CompareDigests returns true when both paths produce the same digest.
func CompareDigests(ctx context.Context, client *api.Client, pathA, pathB string) (bool, error) {
	da, err := DigestSecrets(ctx, client, pathA)
	if err != nil {
		return false, fmt.Errorf("digest A: %w", err)
	}
	db, err := DigestSecrets(ctx, client, pathB)
	if err != nil {
		return false, fmt.Errorf("digest B: %w", err)
	}
	return da.Digest == db.Digest, nil
}

func extractDigestMount(path string) (string, string) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
