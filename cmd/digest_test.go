package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeDigestCmd(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"digest"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestDigestCmd_MissingArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "digest", Args: cobra.RangeArgs(1, 2), RunE: runDigest}
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestDigestCmd_TooManyArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "digest", Args: cobra.RangeArgs(1, 2), RunE: runDigest}
	cmd.SetArgs([]string{"a", "b", "c"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when too many args provided")
	}
}

func TestDigestCmd_UsageText(t *testing.T) {
	use := digestCmd.Use
	if !strings.Contains(use, "digest") {
		t.Errorf("expected 'digest' in Use, got: %s", use)
	}
}

func TestDigestCmd_CompareRequiresTwoPaths(t *testing.T) {
	cmd := &cobra.Command{
		Use:  "digest",
		Args: cobra.RangeArgs(1, 2),
		RunE: runDigest,
	}
	cmd.Flags().BoolVar(&digestCompare, "compare", false, "")
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--compare", "secret/only-one"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --compare used with single path")
	}
	if !strings.Contains(err.Error(), "two paths") {
		t.Errorf("expected 'two paths' in error, got: %v", err)
	}
}

func TestDigestCmd_CompareFlag(t *testing.T) {
	if digestCmd.Flags().Lookup("compare") == nil {
		t.Error("expected --compare flag to be registered")
	}
}
