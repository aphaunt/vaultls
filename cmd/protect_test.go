package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func executeProtectCmd(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestProtectCmd_MissingArgs(t *testing.T) {
	_, err := executeProtectCmd("protect")
	if err == nil {
		t.Fatal("expected error when no path provided")
	}
}

func TestUnprotectCmd_MissingArgs(t *testing.T) {
	_, err := executeProtectCmd("unprotect")
	if err == nil {
		t.Fatal("expected error when no path provided")
	}
}

func TestProtectStatusCmd_MissingArgs(t *testing.T) {
	_, err := executeProtectCmd("protect-status")
	if err == nil {
		t.Fatal("expected error when no path provided")
	}
}

func TestProtectCmd_UsageText(t *testing.T) {
	var cmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Use == "protect <path>" {
			cmd = c
			break
		}
	}
	if cmd == nil {
		t.Fatal("protect command not registered")
	}
	if cmd.Short == "" {
		t.Error("protect command should have a short description")
	}
}

func TestUnprotectCmd_UsageText(t *testing.T) {
	var cmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Use == "unprotect <path>" {
			cmd = c
			break
		}
	}
	if cmd == nil {
		t.Fatal("unprotect command not registered")
	}
	if cmd.Short == "" {
		t.Error("unprotect command should have a short description")
	}
}

func TestProtectStatusCmd_UsageText(t *testing.T) {
	var cmd *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Use == "protect-status <path>" {
			cmd = c
			break
		}
	}
	if cmd == nil {
		t.Fatal("protect-status command not registered")
	}
	if cmd.Short == "" {
		t.Error("protect-status command should have a short description")
	}
}
