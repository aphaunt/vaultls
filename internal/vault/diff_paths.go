package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// DiffPathsResult holds the result of comparing two vault paths.
type DiffPathsResult struct {
	OnlyInA []string
	OnlyInB []string
	InBoth  []string
}

// DiffPaths compares the keys at two vault paths and returns which keys
// are unique to each path and which are shared.
func DiffPaths(ctx context.Context, client *api.Client, pathA, pathB string) (*DiffPathsResult, error) {
	if pathA == "" {
		return nil, fmt.Errorf("pathA must not be empty")
	}
	if pathB == "" {
		return nil, fmt.Errorf("pathB must not be empty")
	}

	keysA, err := listKeys(ctx, client, pathA)
	if err != nil {
		return nil, fmt.Errorf("listing keys at %s: %w", pathA, err)
	}

	keysB, err := listKeys(ctx, client, pathB)
	if err != nil {
		return nil, fmt.Errorf("listing keys at %s: %w", pathB, err)
	}

	setA := toStringSet(keysA)
	setB := toStringSet(keysB)

	result := &DiffPathsResult{}
	for k := range setA {
		if setB[k] {
			result.InBoth = append(result.InBoth, k)
		} else {
			result.OnlyInA = append(result.OnlyInA, k)
		}
	}
	for k := range setB {
		if !setA[k] {
			result.OnlyInB = append(result.OnlyInB, k)
		}
	}
	return result, nil
}

func listKeys(ctx context.Context, client *api.Client, path string) ([]string, error) {
	secret, err := client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}
	raw, ok := secret.Data["keys"]
	if !ok {
		return nil, nil
	}
	iface, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected keys type")
	}
	keys := make([]string, 0, len(iface))
	for _, v := range iface {
		if s, ok := v.(string); ok {
			keys = append(keys, s)
		}
	}
	return keys, nil
}

func toStringSet(keys []string) map[string]bool {
	s := make(map[string]bool, len(keys))
	for _, k := range keys {
		s[k] = true
	}
	return s
}
