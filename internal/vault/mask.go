package vault

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// MaskSecrets replaces secret values matching the given patterns with a redaction string.
// If dryRun is true, changes are printed but not written back to Vault.
func MaskSecrets(ctx context.Context, client *vaultapi.Client, path string, patterns []string, maskWith string, dryRun bool) (int, error) {
	if path == "" {
		return 0, fmt.Errorf("path must not be empty")
	}
	if len(patterns) == 0 {
		return 0, fmt.Errorf("at least one pattern is required")
	}
	if maskWith == "" {
		maskWith = "***"
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return 0, fmt.Errorf("invalid pattern %q: %w", p, err)
		}
		compiled = append(compiled, re)
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return 0, fmt.Errorf("failed to read path %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return 0, fmt.Errorf("no data found at path %q", path)
	}

	data, ok := secret.Data["data"]
	if !ok {
		data = secret.Data
	}
	kv, ok := data.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("unexpected data format at path %q", path)
	}

	masked := make(map[string]interface{}, len(kv))
	count := 0
	for k, v := range kv {
		str, isStr := v.(string)
		if !isStr {
			masked[k] = v
			continue
		}
		original := str
		for _, re := range compiled {
			str = re.ReplaceAllString(str, maskWith)
		}
		if str != original {
			count++
		}
		masked[k] = str
	}

	if count == 0 {
		return 0, nil
	}

	if !dryRun {
		writePath := toMaskDataPath(path)
		_, err = client.Logical().WriteWithContext(ctx, writePath, map[string]interface{}{"data": masked})
		if err != nil {
			return 0, fmt.Errorf("failed to write masked secrets to %q: %w", path, err)
		}
	}

	return count, nil
}

func toMaskDataPath(path string) string {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return path
	}
	return parts[0] + "/data/" + parts[1]
}
