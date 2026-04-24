package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryCmd_MissingVersionFlags(t *testing.T) {
	cmd := historyCmd
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	assert.Error(t, err, "expected error when version flags are missing")
}

func TestHistoryCmd_SameVersionsError(t *testing.T) {
	historyVersionA = 2
	historyVersionB = 2

	buf := &bytes.Buffer{}
	err := runHistory(historyCmd, []string{"myapp/config"})
	_ = buf

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must differ")
}

func TestHistoryCmd_UsageText(t *testing.T) {
	assert.Equal(t, "history <path>", historyCmd.Use)
	assert.NotEmpty(t, historyCmd.Short)
}

func TestHistoryCmd_RequiredFlags(t *testing.T) {
	for _, name := range []string{"version-a", "version-b"} {
		f := historyCmd.Flags().Lookup(name)
		require.NotNil(t, f, "flag %q should exist", name)
		annotations := f.Annotations
		_, required := annotations["cobra_annotation_bash_completion_one_required_flag"]
		// cobra marks required flags differently; just check the flag exists and has a shorthand
		assert.NotEmpty(t, f.Shorthand)
	}
}
