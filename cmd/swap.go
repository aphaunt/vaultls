package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"vaultls/internal/vault"
)

var swapDryRun bool

var swapCmd = &cobra.Command{
	Use:   "swap <pathA> <pathB>",
	Short: "Swap the contents of two Vault secret paths",
	Long: `Swap atomically exchanges the key-value data stored at two KV paths.

Example:
  vaultls swap secret/data/staging secret/data/canary
  vaultls swap --dry-run secret/data/a secret/data/b`,
	Args: cobra.ExactArgs(2),
	RunE: runSwap,
}

func init() {
	swapCmd.Flags().BoolVar(&swapDryRun, "dry-run", false, "Preview the swap without writing changes")
	rootCmd.AddCommand(swapCmd)
}

func runSwap(cmd *cobra.Command, args []string) error {
	pathA := args[0]
	pathB := args[1]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("initialising vault client: %w", err)
	}

	if err := vault.SwapSecrets(cmd.Context(), client, pathA, pathB, swapDryRun); err != nil {
		return fmt.Errorf("swap failed: %w", err)
	}

	if swapDryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would swap %s <-> %s\n", pathA, pathB)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "swapped %s <-> %s\n", pathA, pathB)
	}
	return nil
}
