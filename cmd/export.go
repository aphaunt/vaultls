package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultls/internal/vault"
)

var (
	exportPath   string
	exportFormat string
)

var exportCmd = &cobra.Command{
	Use:   "export <path>",
	Short: "Export secrets from a Vault path to stdout",
	Args:  cobra.ExactArgs(1),
	RunE:  runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Output format: json, yaml, env")
}

func runExport(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	path := args[0]
	secrets, err := client.GetSecrets(path)
	if err != nil {
		return fmt.Errorf("failed to read secrets at %q: %w", path, err)
	}

	fmt := vault.ExportFormat(exportFormat)
	if err := vault.ExportSecrets(os.Stdout, secrets, fmt); err != nil {
		return fmt.Errorf("failed to export secrets: %w", err)
	}
	return nil
}
