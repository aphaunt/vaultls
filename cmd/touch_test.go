package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func executeTouchCmd(args []string) (string, error) {
	root := &cobra.Command{Use: "vaultls"}
	root.AddCommand(touchCmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err := root.Execute()
	return buf.String(), err
}

// TestTouchCmd_MissingArgs verifies that touch requires at least one path argument.
func TestTouchCmd_MissingArgs(t *testing.T) {
	_, err := executeTouchCmd([]string{"touch"})
	if err == nil {
		t.Fatal("expected error when no path provided, got nil")
	}
}

// TestTouchCmd_UsageText verifies that the touch command has a non-empty usage string.
func TestTouchCmd_UsageText(t *testing.T) {
	if touchCmd.Use == "" {
		t.Error("expected touchCmd.Use to be non-empty")
	}
	if touchCmd.Short == "" {
		t.Error("expected touchCmd.Short to be non-empty")
	}
}

// TestTouchCmd_DryRunFlag verifies that the --dry-run flag is registered on the touch command.
func TestTouchCmd_DryRunFlag(t *testing.T) {
	flag := touchCmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Fatal("expected --dry-run flag to be registered on touch command")
	}
	if flag.DefValue != "false" {
		t.Errorf("expected --dry-run default to be false, got %q", flag.DefValue)
	}
}

// TestTouchCmd_RecursiveFlag verifies that the --recursive flag is registered on the touch command.
func TestTouchCmd_RecursiveFlag(t *testing.T) {
	flag := touchCmd.Flags().Lookup("recursive")
	if flag == nil {
		t.Fatal("expected --recursive flag to be registered on touch command")
	}
	if flag.DefValue != "false" {
		t.Errorf("expected --recursive default to be false, got %q", flag.DefValue)
	}
}
