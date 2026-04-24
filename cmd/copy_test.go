package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func executeCopyCmd(args []string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestCopyCmd_MissingArgs(t *testing.T) {
	_, err := executeCopyCmd([]string{"copy"})
	if err == nil {
		t.Fatal("expected error for missing arguments")
	}
}

func TestCopyCmd_SamePathError(t *testing.T) {
	_, err := executeCopyCmd([]string{"copy", "secret/data/foo", "secret/data/foo"})
	if err == nil {
		t.Fatal("expected error when src and dst are the same")
	}
}

func TestCopyCmd_UsageText(t *testing.T) {
	var cmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Use == "copy <src-path> <dst-path>" {
			cmd = c
			break
		}
	}
	if cmd == nil {
		t.Fatal("copy command not registered")
	}
	if cmd.Short == "" {
		t.Error("copy command should have a short description")
	}
}

func TestCopyCmd_OverwriteFlag(t *testing.T) {
	var cmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Use == "copy <src-path> <dst-path>" {
			cmd = c
			break
		}
	}
	if cmd == nil {
		t.Fatal("copy command not registered")
	}
	if cmd.Flags().Lookup("overwrite") == nil {
		t.Error("expected --overwrite flag to be defined")
	}
}
