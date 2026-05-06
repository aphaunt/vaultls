package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func executeScaffoldCmd(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"scaffold"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestScaffoldCmd_MissingArgs(t *testing.T) {
	_, err := executeScaffoldCmd()
	if err == nil {
		t.Fatal("expected error when path is missing")
	}
}

func TestScaffoldCmd_MissingKeysFlag(t *testing.T) {
	_, err := executeScaffoldCmd("secret/app")
	if err == nil {
		t.Fatal("expected error when --keys flag is missing")
	}
}

func TestScaffoldCmd_UsageText(t *testing.T) {
	cmd := &cobra.Command{}
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "scaffold" {
			cmd = sub
			break
		}
	}
	if cmd.Use == "" {
		t.Fatal("scaffold command not registered")
	}
	if cmd.Short == "" {
		t.Fatal("scaffold command missing short description")
	}
}

func TestScaffoldCmd_InvalidDefaultFormat(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	t.Setenv("VAULT_TOKEN", "test")
	_, err := executeScaffoldCmd("secret/app", "--keys", "FOO", "--defaults", "BADFORMAT")
	if err == nil {
		t.Fatal("expected error for invalid default pair format")
	}
}

func TestScaffoldCmd_DryRunFlag(t *testing.T) {
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "scaffold" {
			if f := sub.Flags().Lookup("dry-run"); f == nil {
				t.Fatal("expected --dry-run flag on scaffold command")
			}
			return
		}
	}
	t.Fatal("scaffold command not found")
}
