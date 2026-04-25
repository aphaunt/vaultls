package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func executeRenameCmd(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"rename"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestRenameCmd_MissingArgs(t *testing.T) {
	_, err := executeRenameCmd()
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestRenameCmd_TooFewArgs(t *testing.T) {
	_, err := executeRenameCmd("only-one-arg")
	if err == nil {
		t.Fatal("expected error with only one argument")
	}
}

func TestRenameCmd_UsageText(t *testing.T) {
	cmd := &cobra.Command{}
	renameCmd.SetOut(cmd.OutOrStdout())
	use := renameCmd.Use
	if use == "" {
		t.Fatal("rename command should have a Use field")
	}
	expected := "rename <source> <destination>"
	if use != expected {
		t.Errorf("expected Use %q, got %q", expected, use)
	}
}

func TestRenameCmd_OverwriteFlag(t *testing.T) {
	flag := renameCmd.Flags().Lookup("overwrite")
	if flag == nil {
		t.Fatal("expected --overwrite flag to be defined")
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default value 'false', got %q", flag.DefValue)
	}
}
