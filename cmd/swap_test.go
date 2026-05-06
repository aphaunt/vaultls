package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func executeSwapCmd(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"swap"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestSwapCmd_MissingArgs(t *testing.T) {
	_, err := executeSwapCmd()
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestSwapCmd_TooFewArgs(t *testing.T) {
	_, err := executeSwapCmd("secret/data/only-one")
	if err == nil {
		t.Fatal("expected error with only one argument")
	}
}

func TestSwapCmd_UsageText(t *testing.T) {
	cmd := &cobra.Command{}
	swapCmd.SetOut(cmd.OutOrStdout())
	use := swapCmd.Use
	if use == "" {
		t.Fatal("expected non-empty Use field")
	}
	if swapCmd.Short == "" {
		t.Fatal("expected non-empty Short description")
	}
}

func TestSwapCmd_DryRunFlag(t *testing.T) {
	f := swapCmd.Flags().Lookup("dry-run")
	if f == nil {
		t.Fatal("expected --dry-run flag to be registered")
	}
	if f.DefValue != "false" {
		t.Errorf("expected default false, got %s", f.DefValue)
	}
}

func TestSwapCmd_ExactArgs(t *testing.T) {
	_, err := executeSwapCmd("a", "b", "c")
	if err == nil {
		t.Fatal("expected error with three arguments")
	}
}
