package vault

import (
	"context"
	"fmt"
	"sort"
)

// CompareResult holds the comparison between two vault paths.
type CompareResult struct {
	OnlyInA  []string
	OnlyInB  []string
	Changed  map[string][2]string // key -> [valueA, valueB]
	Identical []string
}

// CompareSecrets performs a full key+value comparison between two secret paths.
func CompareSecrets(ctx context.Context, client *Client, pathA, pathB string) (*CompareResult, error) {
	secretsA, err := client.Read(ctx, pathA)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", pathA, err)
	}

	secretsB, err := client.Read(ctx, pathB)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", pathB, err)
	}

	result := &CompareResult{
		Changed: make(map[string][2]string),
	}

	setA := make(map[string]string)
	setB := make(map[string]string)

	for k, v := range secretsA {
		setA[k] = fmt.Sprintf("%v", v)
	}
	for k, v := range secretsB {
		setB[k] = fmt.Sprintf("%v", v)
	}

	for k, vA := range setA {
		vB, exists := setB[k]
		if !exists {
			result.OnlyInA = append(result.OnlyInA, k)
		} else if vA != vB {
			result.Changed[k] = [2]string{vA, vB}
		} else {
			result.Identical = append(result.Identical, k)
		}
	}

	for k := range setB {
		if _, exists := setA[k]; !exists {
			result.OnlyInB = append(result.OnlyInB, k)
		}
	}

	sort.Strings(result.OnlyInA)
	sort.Strings(result.OnlyInB)
	sort.Strings(result.Identical)

	return result, nil
}
