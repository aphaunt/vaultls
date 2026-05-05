package vault

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// RedactOptions configures redaction behaviour.
type RedactOptions struct {
	Patterns []string // regex patterns matching keys whose values should be redacted
	Placeholder string  // replacement string, defaults to "***REDACTED***"
	DryRun   bool
}

// RedactSecrets reads secrets at path, replaces values of keys matching any
// pattern with the placeholder, and writes the result back unless DryRun is set.
func RedactSecrets(ctx context.Context, client *vaultapi.Client, path string, opts RedactOptions) (map[string]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if len(opts.Patterns) == 0 {
		return nil, fmt.Errorf("at least one pattern is required")
	}
	if opts.Placeholder == "" {
		opts.Placeholder = "***REDACTED***"
	}

	compiledPatterns := make([]*regexp.Regexp, 0, len(opts.Patterns))
	for _, p := range opts.Patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", p, err)
		}
		compiledPatterns = append(compiledPatterns, re)
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at %s", path)
	}

	raw, _ := secret.Data["data"].(map[string]interface{})
	if raw == nil {
		raw = secret.Data
	}

	result := make(map[string]string, len(raw))
	redacted := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		val := fmt.Sprintf("%v", v)
		if matchesPatterns(k, compiledPatterns) {
			result[k] = opts.Placeholder
			redacted[k] = opts.Placeholder
		} else {
			result[k] = val
			redacted[k] = val
		}
	}

	if !opts.DryRun {
		writePath := toDataPath(path)
		_, err = client.Logical().WriteWithContext(ctx, writePath, map[string]interface{}{"data": redacted})
		if err != nil {
			return nil, fmt.Errorf("writing redacted secrets to %s: %w", writePath, err)
		}
	}

	return result, nil
}

func matchesPatterns(key string, patterns []*regexp.Regexp) bool {
	for _, re := range patterns {
		if re.MatchString(key) {
			return true
		}
	}
	return false
}

func toDataPath(path string) string {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 {
		return parts[0] + "/data/" + parts[1]
	}
	return path
}
