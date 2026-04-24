package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func executeAuditCmd(args []string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"audit"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestAuditCmd_MissingPath(t *testing.T) {
	cmd := &cobra.Command{
		Use:  "audit <path>",
		Args: cobra.ExactArgs(1),
		RunE: runAudit,
	}
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing path argument, got nil")
	}
}

func TestAuditCmd_UsageText(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "audit <path>",
		Short: "Show recent audit log entries for a Vault path",
		Args:  cobra.ExactArgs(1),
		RunE:  runAudit,
	}
	usage := cmd.UsageString()
	if usage == "" {
		t.Error("expected non-empty usage string")
	}
}

func TestAuditCmd_OperationFlag(t *testing.T) {
	cmd := &cobra.Command{
		Use:  "audit <path>",
		Args: cobra.ExactArgs(1),
		RunE: runAudit,
	}
	var op string
	cmd.Flags().StringVarP(&op, "operation", "o", "", "Filter by operation")

	if err := cmd.Flags().Set("operation", "read"); err != nil {
		t.Fatalf("failed to set operation flag: %v", err)
	}
	val, err := cmd.Flags().GetString("operation")
	if err != nil {
		t.Fatalf("failed to get operation flag: %v", err)
	}
	if val != "read" {
		t.Errorf("expected operation=read, got %q", val)
	}
}
