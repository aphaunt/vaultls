package vault

import (
	"context"
	"fmt"
	"strings"
)

// ValidationResult holds the outcome of validating a secret path.
type ValidationResult struct {
	Path    string
	Missing []string
	Empty   []string
	Valid   []string
}

// ValidateSecrets checks that all keys at the given path are present and non-empty.
func ValidateSecrets(ctx context.Context, client *Client, path string, requiredKeys []string) (*ValidationResult, error) {
	secrets, err := client.GetSecrets(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets at %q: %w", path, err)
	}

	result := &ValidationResult{Path: path}

	keySet := make(map[string]string, len(secrets))
	for k, v := range secrets {
		keySet[k] = v
	}

	for _, key := range requiredKeys {
		val, exists := keySet[key]
		if !exists {
			result.Missing = append(result.Missing, key)
		} else if strings.TrimSpace(val) == "" {
			result.Empty = append(result.Empty, key)
		} else {
			result.Valid = append(result.Valid, key)
		}
	}

	return result, nil
}

// IsValid returns true if there are no missing or empty keys.
func (r *ValidationResult) IsValid() bool {
	return len(r.Missing) == 0 && len(r.Empty) == 0
}
