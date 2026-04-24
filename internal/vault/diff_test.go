package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffKeys_DisjointSets(t *testing.T) {
	result := DiffKeys([]string{"a", "b"}, []string{"c", "d"})
	assert.Equal(t, []string{"a", "b"}, result.OnlyInA)
	assert.Equal(t, []string{"c", "d"}, result.OnlyInB)
	assert.Empty(t, result.InBoth)
}

func TestDiffKeys_OverlappingSets(t *testing.T) {
	result := DiffKeys([]string{"a", "b", "c"}, []string{"b", "c", "d"})
	assert.Equal(t, []string{"a"}, result.OnlyInA)
	assert.Equal(t, []string{"d"}, result.OnlyInB)
	assert.Equal(t, []string{"b", "c"}, result.InBoth)
}

func TestDiffKeys_IdenticalSets(t *testing.T) {
	result := DiffKeys([]string{"x", "y"}, []string{"x", "y"})
	assert.Empty(t, result.OnlyInA)
	assert.Empty(t, result.OnlyInB)
	assert.Equal(t, []string{"x", "y"}, result.InBoth)
}

func TestDiffSecrets_DetectsChanges(t *testing.T) {
	secretsA := map[string]string{"foo": "bar", "baz": "qux", "only_a": "val"}
	secretsB := map[string]string{"foo": "changed", "baz": "qux", "only_b": "val"}

	result := DiffSecrets(secretsA, secretsB)

	assert.Equal(t, []string{"only_a"}, result.OnlyInA)
	assert.Equal(t, []string{"only_b"}, result.OnlyInB)
	assert.Contains(t, result.InBoth, "foo")
	assert.Contains(t, result.InBoth, "baz")
	assert.Equal(t, [2]string{"bar", "changed"}, result.Changed["foo"])
	assert.NotContains(t, result.Changed, "baz")
}

func TestDiffSecrets_NoChanges(t *testing.T) {
	secrets := map[string]string{"a": "1", "b": "2"}
	result := DiffSecrets(secrets, secrets)
	assert.Empty(t, result.OnlyInA)
	assert.Empty(t, result.OnlyInB)
	assert.Empty(t, result.Changed)
}
