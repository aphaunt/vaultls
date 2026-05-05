package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func executeRevertCmd(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"revert"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestRevertCmd_MissingArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "revert", Args: cobra.ExactArgs(2), RunE: runRevert}
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestRevertCmd_TooFewArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "revert", Args: cobra.ExactArgs(2), RunE: runRevert}
	cmd.SetArgs([]string{"secret/myapp"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when only one arg provided")
	}
}

func TestRevertCmd_InvalidVersion(t *testing.T) {
	cmd := &cobra.Command{Use: "revert", Args: cobra.ExactArgs(2), RunE: runRevert}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"secret/myapp", "notanumber"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-integer version")
	}
}

func TestRevertCmd_UsageText(t *testing.T) {
	cmd := &cobra.Command{Use: "revert", Args: cobra.ExactArgs(2), RunE: runRevert}
	if cmd.Use != "revert" {
		t.Errorf("expected use 'revert', got %q", cmd.Use)
	}
}

func TestRevertCmd_DryRunFlag(t *testing.T) {
	var dryRun bool
	cmd := &cobra.Command{Use: "revert", Args: cobra.ExactArgs(2), RunE: func(c *cobra.Command, args []string) error {
		dryRun, _ = c.Flags().GetBool("dry-run")
		return fmt.Errorf("stop")
	}}
	cmd.Flags().Bool("dry-run", false, "")
	cmd.SetArgs([]string{"secret/myapp", "1", "--dry-run"})
	_ = cmd.Execute()
	if !dryRun {
		t.Error("expected dry-run flag to be true")
	}
}
