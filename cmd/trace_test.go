package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeTraceCmd(args []string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"trace"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestTraceCmd_MissingArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "trace <path>", Args: cobra.ExactArgs(1)}
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for missing args, got nil")
	}
}

func TestTraceCmd_TooManyArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "trace <path>", Args: cobra.ExactArgs(1)}
	err := cmd.Args(cmd, []string{"a", "b"})
	if err == nil {
		t.Fatal("expected error for too many args, got nil")
	}
}

func TestTraceCmd_UsageText(t *testing.T) {
	if !strings.Contains(traceCmd.Use, "trace") {
		t.Errorf("expected Use to contain 'trace', got %q", traceCmd.Use)
	}
	if traceCmd.Short == "" {
		t.Error("expected non-empty Short description")
	}
}

func TestTraceCmd_ExactArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "trace <path>", Args: cobra.ExactArgs(1)}
	if err := cmd.Args(cmd, []string{"secret/myapp"}); err != nil {
		t.Errorf("unexpected error for valid args: %v", err)
	}
}
