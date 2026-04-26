package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultls/internal/vault"
)

// touchCmd represents the touch command, which re-writes secrets at a given
// path without changing their values. This forces a new version to be created
// in Vault KV v2, which is useful for triggering downstream watchers or
// auditing purposes.
var touchCmd = &cobra.Command{
	Use:   "touch <path>",
	Short: "Re-write secrets at a path to create a new version",
	Long: `Touch reads the current secrets at the given Vault KV v2 path and
writes them back unchanged, creating a new version entry.

This is useful when you need to:
  - Trigger secret rotation watchers
  - Bump the version metadata without changing values
  - Verify write access to a path

Example:
  vaultls touch secret/myapp/config
  vaultls touch secret/myapp/config --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runTouch,
}

func init() {
	rootCmd.AddCommand(touchCmd)
	touchCmd.Flags().Bool("dry-run", false, "Print what would be written without making changes")
}

func runTouch(cmd *cobra.Command, args []string) error {
	path := args[0]
	if path == "" {
		return fmt.Errorf("path must not be empty")
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fmt.Errorf("failed to read dry-run flag: %w", err)
	}

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	result, err := vault.TouchSecrets(cmd.Context(), client, path, dryRun)
	if err != nil {
		return fmt.Errorf("touch failed: %w", err)
	}

	if dryRun {
		fmt.Fprintf(os.Stdout, "[dry-run] would touch %d key(s) at %s\n", result.KeyCount, path)
	} else {
		fmt.Fprintf(os.Stdout, "touched %d key(s) at %s (new version: %d)\n", result.KeyCount, path, result.NewVersion)
	}

	return nil
}
