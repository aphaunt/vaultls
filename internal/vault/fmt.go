package vault

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/vault/api"
)

// FormatResult holds the outcome of a format operation on a single key.
type FormatResult struct {
	Key     string
	OldVal  string
	NewVal  string
	Changed bool
}

// FmtSecrets reads all secrets at path, normalises key names to the
// requested style ("upper" or "lower"), and writes them back.
// It returns one FormatResult per key so callers can report changes.
func FmtSecrets(ctx context.Context, client *api.Client, path, style string, dryRun bool) ([]FormatResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if style != "upper" && style != "lower" {
		return nil, fmt.Errorf("style must be \"upper\" or \"lower\", got %q", style)
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at %s", path)
	}

	data, ok := secret.Data["data"]
	if !ok {
		data = secret.Data
	}
	raw, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data shape at %s", path)
	}

	newData := make(map[string]interface{}, len(raw))
	results := make([]FormatResult, 0, len(raw))

	keys := make([]string, 0, len(raw))
	for k := range raw {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := raw[k]
		var newKey string
		if style == "upper" {
			newKey = strings.ToUpper(k)
		} else {
			newKey = strings.ToLower(k)
		}
		newData[newKey] = v
		results = append(results, FormatResult{
			Key:     k,
			OldVal:  k,
			NewVal:  newKey,
			Changed: newKey != k,
		})
	}

	if !dryRun {
		writePayload := map[string]interface{}{"data": newData}
		_, err = client.Logical().WriteWithContext(ctx, path, writePayload)
		if err != nil {
			return nil, fmt.Errorf("writing %s: %w", path, err)
		}
	}

	return results, nil
}
