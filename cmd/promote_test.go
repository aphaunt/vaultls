package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executePromoteCmd(args []string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestPromoteCmd_MissingArgs(t *testing.T) {
	_, err := executePromoteCmd([]string{"promote"})
	if err == nil {
		t.Fatal("expected error for missing arguments")
	}
}

func TestPromoteCmd_TooFewArgs(t *testing.T) {
	_, err := executePromoteCmd([]string{"promote", "secret/dev"})
	if err == nil {
		t.Fatal("expected error for too few arguments")
	}
}

func TestPromoteCmd_UsageText(t *testing.T) {
	cmd := &cobra.Command{}
	promoteCmd.SetOut(cmd.OutOrStdout())
	use := promoteCmd.Use
	if !strings.HasPrefix(use, "promote") {
		t.Errorf("unexpected Use string: %q", use)
	}
}

func TestPromoteCmd_OverwriteFlag(t *testing.T) {
	f := promoteCmd.Flags().Lookup("overwrite")
	if f == nil {
		t.Fatal("expected --overwrite flag to be defined")
	}
	if f.DefValue != "false" {
		t.Errorf("expected default false, got %q", f.DefValue)
	}
}

func TestPromoteCmd_SamePathError(t *testing.T) {
	_, err := executePromoteCmd([]string{
		"promote",
		"--address", "http://127.0.0.1:1",
		"--token", "tok",
		"secret/dev", "secret/dev",
	})
	if err == nil {
		t.Fatal("expected error when src and dst are the same")
	}
}
