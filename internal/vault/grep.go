package vault

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// GrepResult holds a single match found during a grep operation.
type GrepResult struct {
	Path  string
	Key   string
	Value string
}

// GrepSecrets searches all secrets under the given path for keys or values
// matching the provided regular expression pattern.
func GrepSecrets(ctx context.Context, client *vaultapi.Client, path, pattern string, keysOnly bool) ([]GrepResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if pattern == "" {
		return nil, fmt.Errorf("pattern must not be empty")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}

	listPath := fmt.Sprintf("%s/metadata/%s", extractMount(path), extractSubPath(path))
	secret, err := client.Logical().ListWithContext(ctx, listPath)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}

	var results []GrepResult
	for _, k := range keys {
		key, _ := k.(string)
		if strings.HasSuffix(key, "/") {
			continue
		}
		readPath := fmt.Sprintf("%s/data/%s/%s", extractMount(path), extractSubPath(path), key)
		s, err := client.Logical().ReadWithContext(ctx, readPath)
		if err != nil || s == nil || s.Data == nil {
			continue
		}
		data, _ := s.Data["data"].(map[string]interface{})
		for field, val := range data {
			strVal := fmt.Sprintf("%v", val)
			if re.MatchString(field) || (!keysOnly && re.MatchString(strVal)) {
				results = append(results, GrepResult{
					Path:  path + "/" + key,
					Key:   field,
					Value: strVal,
				})
			}
		}
	}
	return results, nil
}

func extractMount(path string) string {
	parts := strings.SplitN(strings.Trim(path, "/"), "/", 2)
	return parts[0]
}

func extractSubPath(path string) string {
	parts := strings.SplitN(strings.Trim(path, "/"), "/", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
