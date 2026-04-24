package vault

import (
	"context"
	"fmt"
	"strings"
)

// SearchResult holds a matching secret path and the matching key/value pairs.
type SearchResult struct {
	Path    string
	Matches map[string]string
}

// SearchSecrets recursively searches all secrets under the given root path
// for keys or values containing the query string (case-insensitive).
func SearchSecrets(ctx context.Context, client *Client, root, query string) ([]SearchResult, error) {
	keys, err := client.List(ctx, root)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", root, err)
	}

	var results []SearchResult
	lower := strings.ToLower(query)

	for _, key := range keys {
		fullPath := strings.TrimRight(root, "/") + "/" + key

		if strings.HasSuffix(key, "/") {
			// Recurse into sub-path
			sub, err := SearchSecrets(ctx, client, fullPath, query)
			if err != nil {
				return nil, err
			}
			results = append(results, sub...)
			continue
		}

		secrets, err := client.Read(ctx, fullPath)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", fullPath, err)
		}

		matches := make(map[string]string)
		for k, v := range secrets {
			if strings.Contains(strings.ToLower(k), lower) ||
				strings.Contains(strings.ToLower(v), lower) {
				matches[k] = v
			}
		}

		if len(matches) > 0 {
			results = append(results, SearchResult{
				Path:    fullPath,
				Matches: matches,
			})
		}
	}

	return results, nil
}
