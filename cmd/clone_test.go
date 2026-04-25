package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeCloneCmd(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestCloneCmd_MissingArgs(t *testing.T) {
	_, err := executeCloneCmd("clone")
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestCloneCmd_TooFewArgs(t *testing.T) {
	_, err := executeCloneCmd("clone", "only-one-arg")
	if err == nil {
		t.Fatal("expected error when only one arg provided")
	}
}

func TestCloneCmd_UsageText(t *testing.T) {
	cmd := &cobra.Command{}
	cloneCmd.SetOut(cmd.OutOrStdout())
	use := cloneCmd.Use
	if !strings.Contains(use, "src-path") || !strings.Contains(use, "dst-path") {
		t.Errorf("unexpected usage string: %q", use)
	}
}

func TestCloneCmd_OverwriteFlag(t *testing.T) {
	f := cloneCmd.Flags().Lookup("overwrite")
	if f == nil {
		t.Fatal("expected --overwrite flag to be defined")
	}
	if f.DefValue != "false" {
		t.Errorf("expected default false, got %q", f.DefValue)
	}
}

func TestCloneCmd_DryRunFlag(t *testing.T) {
	f := cloneCmd.Flags().Lookup("dry-run")
	if f == nil {
		t.Fatal("expected --dry-run flag to be defined")
	}
	if f.DefValue != "false" {
		t.Errorf("expected default false, got %q", f.DefValue)
	}
}
