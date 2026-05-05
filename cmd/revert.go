package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var revertDryRun bool

var revertCmd = &cobra.Command{
	Use:   "revert <path> <version>",
	Short: "Revert a secret to a previous version",
	Long: `Revert reads the specified historical version of a KV v2 secret
and writes it back as the latest version, effectively restoring it.

Use --dry-run to preview the data that would be restored without writing.`,
	Args:    cobra.ExactArgs(2),
	RunE:    runRevert,
	Example: `  vaultls revert secret/myapp 3\n  vaultls revert secret/myapp 2 --dry-run`,
}

func init() {
	rootCmd.AddCommand(revertCmd)
	revertCmd.Flags().BoolVar(&revertDryRun, "dry-run", false, "Preview the revert without writing")
}

func runRevert(cmd *cobra.Command, args []string) error {
	path := args[0]
	versionStr := args[1]

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version %q: must be a positive integer", versionStr)
	}

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("initializing vault client: %w", err)
	}

	data, err := vault.RevertSecrets(cmd.Context(), client, path, version, revertDryRun)
	if err != nil {
		return fmt.Errorf("reverting %s to version %d: %w", path, version, err)
	}

	if revertDryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would restore %s to version %d:\n", path, version)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Reverted %s to version %d:\n", path, version)
	}

	for k, v := range data {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s = %v\n", k, v)
	}
	return nil
}
