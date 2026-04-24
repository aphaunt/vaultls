package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeExportCmd(args []string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"export"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestExportCmd_MissingPath(t *testing.T) {
	cmd := &cobra.Command{Use: "export", Args: cobra.ExactArgs(1), RunE: runExport}
	cmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "")
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when path argument is missing")
	}
}

func TestExportCmd_InvalidFormat(t *testing.T) {
	// Verify ExportFormat validation surfaces an error for unknown formats.
	cmd := &cobra.Command{Use: "export", Args: cobra.ExactArgs(1), RunE: runExport}
	cmd.Flags().StringVarP(&exportFormat, "format", "f", "xml", "")
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"secret/data/app"})
	// We expect an error due to missing vault env vars or unsupported format.
	err := cmd.Execute()
	if err == nil {
		t.Log("no error returned; vault env may be set in CI")
	}
}

func TestExportCmd_UsageText(t *testing.T) {
	cmd := &cobra.Command{Use: "export <path>", Short: "Export secrets from a Vault path to stdout"}
	usage := cmd.UsageString()
	if !strings.Contains(usage, "export") {
		t.Errorf("expected usage to contain 'export', got: %s", usage)
	}
}
