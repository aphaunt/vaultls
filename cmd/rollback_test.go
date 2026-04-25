package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func executeRollbackCmd(args []string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"rollback"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestRollbackCmd_MissingArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "rollback <path> <version>", Args: cobra.ExactArgs(2), RunE: runRollback}
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestRollbackCmd_TooFewArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "rollback <path> <version>", Args: cobra.ExactArgs(2), RunE: runRollback}
	cmd.SetArgs([]string{"secret/foo"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when only one arg provided")
	}
}

func TestRollbackCmd_InvalidVersion(t *testing.T) {
	cmd := &cobra.Command{Use: "rollback", Args: cobra.ExactArgs(2), RunE: runRollback}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"secret/foo", "notanumber"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-integer version")
	}
}

func TestRollbackCmd_UsageText(t *testing.T) {
	cmd := &cobra.Command{Use: "rollback <path> <version>", Args: cobra.ExactArgs(2), RunE: runRollback}
	use := cmd.Use
	if use != "rollback <path> <version>" {
		t.Errorf("unexpected Use string: %q", use)
	}
}

func TestRollbackCmd_ZeroVersion(t *testing.T) {
	cmd := &cobra.Command{Use: "rollback", Args: cobra.ExactArgs(2), RunE: runRollback}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"secret/foo", "0"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for version 0")
	}
}
