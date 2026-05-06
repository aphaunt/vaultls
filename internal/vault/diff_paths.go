package vault

import (
	"context"
	"fmt"
	"sort"

	vaultapi "github.com/hashicorp/vault/api"
)

// PathDiff represents the structural difference between two Vault paths.
type PathDiff struct {
	OnlyInA  []string
	OnlyInB  []string
	InBoth   []string
}

// DiffPaths compares the list of keys at two Vault paths and returns a PathDiff.
func DiffPaths(ctx context.Context, client *vaultapi.Client, pathA, pathB string) (*PathDiff, error) {
	if pathA == "" || pathB == "" {
		return nil, fmt.Errorf("both paths must be non-empty")
	}

	keysA, err := listKeys(ctx, client, pathA)
	if err != nil {
		return nil, fmt.Errorf("listing path %q: %w", pathA, err)
	}

	keysB, err := listKeys(ctx, client, pathB)
	if err != nil {
		return nil, fmt.Errorf("listing path %q: %w", pathB, err)
	}

	setA := toStringSet(keysA)
	setB := toStringSet(keysB)

	diff := &PathDiff{}

	for k := range setA {
		if setB[k] {
			diff.InBoth = append(diff.InBoth, k)
		} else {
			diff.OnlyInA = append(diff.OnlyInA, k)
		}
	}
	for k := range setB {
		if !setA[k] {
			diff.OnlyInB = append(diff.OnlyInB, k)
		}
	}

	sort.Strings(diff.OnlyInA)
	sort.Strings(diff.OnlyInB)
	sort.Strings(diff.InBoth)

	return diff, nil
}

func listKeys(ctx context.Context, client *vaultapi.Client, path string) ([]string, error) {
	secret, err := client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}
	raw, ok := secret.Data["keys"]
	if !ok {
		return []string{}, nil
	}
	ifaces, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected keys format at %q", path)
	}
	keys := make([]string, 0, len(ifaces))
	for _, v := range ifaces {
		if s, ok := v.(string); ok {
			keys = append(keys, s)
		}
	}
	return keys, nil
}

func toStringSet(keys []string) map[string]bool {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}
