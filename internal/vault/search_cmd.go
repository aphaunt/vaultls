package vault

import (
	"context"
	"fmt"
	"strings"
)

// SearchResult holds a matched secret path and the matching key/value pairs.
type SearchResult struct {
	Path    string
	Matches map[string]string
}

// SearchSecretsByValue searches all secrets under basePath for a given value substring.
func SearchSecretsByValue(ctx context.Context, client *Client, basePath, query string) ([]SearchResult, error) {
	keys, err := client.List(ctx, basePath)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", basePath, err)
	}

	var results []SearchResult
	for _, key := range keys {
		fullPath := strings.TrimRight(basePath, "/") + "/" + key
		secrets, err := client.Read(ctx, fullPath)
		if err != nil {
			continue
		}
		matches := make(map[string]string)
		for k, v := range secrets {
			if strings.Contains(v, query) {
				matches[k] = v
			}
		}
		if len(matches) > 0 {
			results = append(results, SearchResult{Path: fullPath, Matches: matches})
		}
	}
	return results, nil
}
