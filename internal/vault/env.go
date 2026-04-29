package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

// EnvDiff represents the difference between two environment paths.
type EnvDiff struct {
	OnlyInA  map[string]string
	OnlyInB  map[string]string
	Changed  map[string][2]string
	Identical map[string]string
}

// CompareEnvs reads secrets from two paths and returns a structured diff
// suitable for environment-oriented output.
func CompareEnvs(ctx context.Context, client *api.Client, pathA, pathB string) (*EnvDiff, error) {
	if pathA == "" || pathB == "" {
		return nil, fmt.Errorf("both pathA and pathB must be non-empty")
	}

	secretsA, err := readKVSecrets(ctx, client, pathA)
	if err != nil {
		return nil, fmt.Errorf("reading path %q: %w", pathA, err)
	}

	secretsB, err := readKVSecrets(ctx, client, pathB)
	if err != nil {
		return nil, fmt.Errorf("reading path %q: %w", pathB, err)
	}

	return diffEnvMaps(secretsA, secretsB), nil
}

func readKVSecrets(ctx context.Context, client *api.Client, path string) (map[string]string, error) {
	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return map[string]string{}, nil
	}

	data, ok := secret.Data["data"]
	if !ok {
		data = secret.Data
	}

	raw, ok := data.(map[string]interface{})
	if !ok {
		return map[string]string{}, nil
	}

	result := make(map[string]string, len(raw))
	for k, v := range raw {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result, nil
}

func diffEnvMaps(a, b map[string]string) *EnvDiff {
	diff := &EnvDiff{
		OnlyInA:   make(map[string]string),
		OnlyInB:   make(map[string]string),
		Changed:   make(map[string][2]string),
		Identical: make(map[string]string),
	}

	for k, va := range a {
		if vb, ok := b[k]; ok {
			if va == vb {
				diff.Identical[k] = va
			} else {
				diff.Changed[k] = [2]string{va, vb}
			}
		} else {
			diff.OnlyInA[k] = va
		}
	}

	for k, vb := range b {
		if _, ok := a[k]; !ok {
			diff.OnlyInB[k] = vb
		}
	}

	return diff
}

// FormatEnvDiff returns a human-readable string of the EnvDiff.
func FormatEnvDiff(d *EnvDiff, pathA, pathB string) string {
	var sb strings.Builder

	for k, v := range d.OnlyInA {
		fmt.Fprintf(&sb, "- [%s only] %s=%s\n", pathA, k, v)
	}
	for k, v := range d.OnlyInB {
		fmt.Fprintf(&sb, "+ [%s only] %s=%s\n", pathB, k, v)
	}
	for k, vs := range d.Changed {
		fmt.Fprintf(&sb, "~ %s: %s -> %s\n", k, vs[0], vs[1])
	}

	if sb.Len() == 0 {
		return "No differences found.\n"
	}
	return sb.String()
}
