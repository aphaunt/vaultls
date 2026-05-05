package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func executeSetCmd(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"set"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestSetCmd_MissingArgs(t *testing.T) {
	_, err := executeSetCmd()
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}

func TestSetCmd_TooFewArgs(t *testing.T) {
	_, err := executeSetCmd("secret/myapp")
	if err == nil {
		t.Fatal("expected error when only path provided")
	}
}

func TestSetCmd_InvalidPairFormat(t *testing.T) {
	_, err := executeSetCmd("secret/myapp", "badformat")
	if err == nil {
		t.Fatal("expected error for invalid key=value format")
	}
}

func TestSetCmd_UsageText(t *testing.T) {
	cmd := &cobra.Command{}
	for _, c := range rootCmd.Commands() {
		if c.Name() == "set" {
			cmd = c
			break
		}
	}
	if cmd.Use == "" {
		t.Fatal("set command not registered")
	}
	if cmd.Short == "" {
		t.Fatal("set command missing short description")
	}
}

func TestSetCmd_DryRunFlag(t *testing.T) {
	cmd := &cobra.Command{}
	for _, c := range rootCmd.Commands() {
		if c.Name() == "set" {
			cmd = c
			break
		}
	}
	if cmd.Flags().Lookup("dry-run") == nil {
		t.Fatal("expected --dry-run flag on set command")
	}
}
