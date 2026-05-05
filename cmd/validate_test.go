package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func executeValidateCmd(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestValidateCmd_MissingArgs(t *testing.T) {
	_, err := executeValidateCmd("validate")
	if err == nil {
		t.Error("expected error for missing arguments")
	}
}

func TestValidateCmd_TooFewArgs(t *testing.T) {
	_, err := executeValidateCmd("validate", "secret/data/app")
	if err == nil {
		t.Error("expected error when only one argument is provided")
	}
}

func TestValidateCmd_UsageText(t *testing.T) {
	var cmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Use == "validate <path> <key1,key2,...>" {
			cmd = c
			break
		}
	}
	if cmd == nil {
		t.Fatal("validate command not registered")
	}
	if cmd.Short == "" {
		t.Error("expected non-empty short description")
	}
}

func TestValidateCmd_TooManyArgs(t *testing.T) {
	_, err := executeValidateCmd("validate", "secret/data/app", "KEY1", "extra")
	if err == nil {
		t.Error("expected error for too many arguments")
	}
}

// findValidateCmd is a helper that locates the validate subcommand from rootCmd.
func findValidateCmd(t *testing.T) *cobra.Command {
	t.Helper()
	for _, c := range rootCmd.Commands() {
		if c.Use == "validate <path> <key1,key2,...>" {
			return c
		}
	}
	t.Fatal("validate command not registered")
	return nil
}

func TestValidateCmd_HasExample(t *testing.T) {
	cmd := findValidateCmd(t)
	if cmd.Example == "" {
		t.Error("expected non-empty example for validate command")
	}
}
