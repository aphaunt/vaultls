package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

// PivotResult holds the restructured secrets keyed by a chosen field.
type PivotResult struct {
	Key   string
	Value map[string]string
}

// PivotSecrets reads secrets from path and restructures them so that the value
// of pivotKey becomes the top-level grouping key. Each group contains the
// remaining key/value pairs from that secret.
func PivotSecrets(ctx context.Context, client *api.Client, path, pivotKey string) ([]PivotResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if pivotKey == "" {
		return nil, fmt.Errorf("pivot key must not be empty")
	}

	listPath := fmt.Sprintf("%s/metadata/%s", extractPivotMount(path), extractPivotSubPath(path))
	secret, err := client.Logical().ListWithContext(ctx, listPath)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", listPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no secrets found at %s", path)
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected list response at %s", path)
	}

	var results []PivotResult
	for _, k := range keys {
		key := strings.TrimSuffix(fmt.Sprintf("%v", k), "/")
		dataPath := fmt.Sprintf("%s/data/%s/%s", extractPivotMount(path), extractPivotSubPath(path), key)
		s, err := client.Logical().ReadWithContext(ctx, dataPath)
		if err != nil || s == nil || s.Data == nil {
			continue
		}
		data, _ := s.Data["data"].(map[string]interface{})
		pivotVal, exists := data[pivotKey]
		if !exists {
			continue
		}
		row := map[string]string{}
		for dk, dv := range data {
			if dk != pivotKey {
				row[dk] = fmt.Sprintf("%v", dv)
			}
		}
		results = append(results, PivotResult{
			Key:   fmt.Sprintf("%v", pivotVal),
			Value: row,
		})
	}
	return results, nil
}

func extractPivotMount(path string) string {
	parts := strings.SplitN(strings.Trim(path, "/"), "/", 2)
	if len(parts) == 0 {
		return path
	}
	return parts[0]
}

func extractPivotSubPath(path string) string {
	parts := strings.SplitN(strings.Trim(path, "/"), "/", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
