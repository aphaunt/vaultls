package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var (
	restoreOverwrite bool
	restoreDryRun   bool
	restoreFile     string
)

var restoreCmd = &cobra.Command{
	Use:   "restore <dest-path>",
	Short: "Restore secrets from a JSON snapshot file into Vault",
	Args:  cobra.ExactArgs(1),
	RunE:  runRestore,
}

func init() {
	restoreCmd.Flags().StringVarP(&restoreFile, "file", "f", "", "Path to the JSON snapshot file (required)")
	_ = restoreCmd.MarkFlagRequired("file")
	restoreCmd.Flags().BoolVar(&restoreOverwrite, "overwrite", false, "Overwrite existing secrets")
	restoreCmd.Flags().BoolVar(&restoreDryRun, "dry-run", false, "Preview changes without writing")
	rootCmd.AddCommand(restoreCmd)
}

func runRestore(cmd *cobra.Command, args []string) error {
	destPath := args[0]

	raw, err := os.ReadFile(restoreFile)
	if err != nil {
		return fmt.Errorf("read snapshot file: %w", err)
	}

	// Validate JSON before passing it down.
	var probe map[string]string
	if err := json.Unmarshal(raw, &probe); err != nil {
		return fmt.Errorf("invalid snapshot JSON: %w", err)
	}

	client, err := vault.NewClient()
	if err != nil {
		return err
	}

	restored, err := vault.RestoreSecrets(cmd.Context(), client, destPath, string(raw), restoreOverwrite, restoreDryRun)
	if err != nil {
		return err
	}

	if restoreDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "[dry-run] would restore:")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "restored:")
	}
	for _, p := range restored {
		fmt.Fprintln(cmd.OutOrStdout(), " ", p)
	}
	if len(restored) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "  (nothing to restore)")
	}
	return nil
}
