package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// RotateResult holds the outcome of rotating a single secret key.
type RotateResult struct {
	Key     string
	OldValue string
	NewValue string
	Rotated  bool
}

// RotateSecrets re-writes each key at path with a new value produced by
// generator. If dryRun is true no writes are performed.
func RotateSecrets(
	ctx context.Context,
	client *vaultapi.Client,
	path string,
	generator func(key, current string) (string, error),
	dryRun bool,
) ([]RotateResult, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if generator == nil {
		return nil, fmt.Errorf("generator must not be nil")
	}

	dataPath := toRotateDataPath(path)
	secret, err := client.Logical().ReadWithContext(ctx, dataPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dataPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at %s", path)
	}

	rawData, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at %s", path)
	}

	results := make([]RotateResult, 0, len(rawData))
	updated := make(map[string]interface{}, len(rawData))

	for k, v := range rawData {
		current := fmt.Sprintf("%v", v)
		newVal, err := generator(k, current)
		if err != nil {
			return nil, fmt.Errorf("generate value for key %q: %w", k, err)
		}
		results = append(results, RotateResult{
			Key:      k,
			OldValue: current,
			NewValue: newVal,
			Rotated:  !dryRun,
		})
		updated[k] = newVal
	}

	if !dryRun {
		_, err = client.Logical().WriteWithContext(ctx, dataPath, map[string]interface{}{"data": updated})
		if err != nil {
			return nil, fmt.Errorf("write %s: %w", dataPath, err)
		}
	}

	return results, nil
}

func toRotateDataPath(path string) string {
	parts := strings.SplitN(strings.Trim(path, "/"), "/", 2)
	if len(parts) == 2 {
		return parts[0] + "/data/" + parts[1]
	}
	return path
}
