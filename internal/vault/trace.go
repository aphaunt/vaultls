package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// TraceEntry represents a single access record for a secret path.
type TraceEntry struct {
	Path      string
	Operation string
	Version   int
	Caller    string
}

// TraceResult holds all trace entries for a given path.
type TraceResult struct {
	Path    string
	Entries []TraceEntry
}

// TraceSecrets retrieves metadata access history for a KV v2 secret path.
func TraceSecrets(ctx context.Context, client *vaultapi.Client, path string) (*TraceResult, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid path: expected <mount>/<secret>, got %q", path)
	}
	mount, secret := parts[0], parts[1]
	metaPath := fmt.Sprintf("%s/metadata/%s", mount, secret)

	secret2, err := client.Logical().ReadWithContext(ctx, metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata at %s: %w", metaPath, err)
	}
	if secret2 == nil || secret2.Data == nil {
		return nil, fmt.Errorf("no metadata found at %s", metaPath)
	}

	result := &TraceResult{Path: path}

	versions, ok := secret2.Data["versions"].(map[string]interface{})
	if !ok {
		return result, nil
	}

	for ver, raw := range versions {
		vMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		op := "read"
		if state, ok := vMap["destruction_time"].(string); ok && state != "" {
			op = "deleted"
		}
		createdBy := ""
		if cb, ok := vMap["created_time"].(string); ok {
			createdBy = cb
		}
		var vNum int
		fmt.Sscanf(ver, "%d", &vNum)
		result.Entries = append(result.Entries, TraceEntry{
			Path:      path,
			Operation: op,
			Version:   vNum,
			Caller:    createdBy,
		})
	}

	return result, nil
}
