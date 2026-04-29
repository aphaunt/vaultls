package vault

import (
	"context"
	"fmt"
	"strings"
)

// LintResult holds the result of a lint check for a single secret key.
type LintResult struct {
	Path    string
	Key     string
	Message string
	Severity string // "warn" or "error"
}

// LintOptions configures which lint rules are applied.
type LintOptions struct {
	DisallowEmptyValues bool
	DisallowUppercaseKeys bool
	RequirePrefix       string
}

// LintSecrets reads secrets at the given path and applies lint rules,
// returning a slice of LintResult for any violations found.
func LintSecrets(ctx context.Context, client *Client, path string, opts LintOptions) ([]LintResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	secrets, err := client.GetSecrets(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets at %s: %w", path, err)
	}

	var results []LintResult

	for key, val := range secrets {
		strVal, _ := val.(string)

		if opts.DisallowEmptyValues && strings.TrimSpace(strVal) == "" {
			results = append(results, LintResult{
				Path:     path,
				Key:      key,
				Message:  "value is empty or whitespace",
				Severity: "error",
			})
		}

		if opts.DisallowUppercaseKeys && key == strings.ToUpper(key) && strings.ContainsAny(key, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			results = append(results, LintResult{
				Path:     path,
				Key:      key,
				Message:  "key is all uppercase; prefer lowercase or snake_case",
				Severity: "warn",
			})
		}

		if opts.RequirePrefix != "" && !strings.HasPrefix(key, opts.RequirePrefix) {
			results = append(results, LintResult{
				Path:     path,
				Key:      key,
				Message:  fmt.Sprintf("key does not have required prefix %q", opts.RequirePrefix),
				Severity: "warn",
			})
		}
	}

	return results, nil
}
