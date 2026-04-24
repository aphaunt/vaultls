package vault

import "sort"

// DiffResult represents the difference between two sets of secrets.
type DiffResult struct {
	OnlyInA  []string
	OnlyInB  []string
	InBoth   []string
	Changed  map[string][2]string
}

// DiffKeys compares two slices of keys and returns a DiffResult.
func DiffKeys(keysA, keysB []string) DiffResult {
	setA := toSet(keysA)
	setB := toSet(keysB)

	result := DiffResult{
		Changed: make(map[string][2]string),
	}

	for k := range setA {
		if _, ok := setB[k]; ok {
			result.InBoth = append(result.InBoth, k)
		} else {
			result.OnlyInA = append(result.OnlyInA, k)
		}
	}

	for k := range setB {
		if _, ok := setA[k]; !ok {
			result.OnlyInB = append(result.OnlyInB, k)
		}
	}

	sort.Strings(result.OnlyInA)
	sort.Strings(result.OnlyInB)
	sort.Strings(result.InBoth)

	return result
}

// DiffSecrets compares two maps of secret key/value pairs and records changed values.
func DiffSecrets(secretsA, secretsB map[string]string) DiffResult {
	keysA := mapKeys(secretsA)
	keysB := mapKeys(secretsB)

	result := DiffKeys(keysA, keysB)

	for _, k := range result.InBoth {
		if secretsA[k] != secretsB[k] {
			result.Changed[k] = [2]string{secretsA[k], secretsB[k]}
		}
	}

	return result
}

func toSet(keys []string) map[string]struct{} {
	s := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		s[k] = struct{}{}
	}
	return s
}

func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
