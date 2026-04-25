package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"vaultls/internal/vault"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback <path> <version>",
	Short: "Roll back a secret to a previous version",
	Long:  `Reads the specified version of a KV v2 secret and writes it as the new current version.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runRollback,
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

func runRollback(cmd *cobra.Command, args []string) error {
	path := args[0]
	versionStr := args[1]

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version %q: must be a positive integer", versionStr)
	}

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	result, err := vault.RollbackSecret(cmd.Context(), client, path, version)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	if result.Success {
		fmt.Fprintf(cmd.OutOrStdout(), "Successfully rolled back %q to version %d\n", result.Path, result.Version)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Rollback of %q to version %d failed\n", result.Path, result.Version)
	}
	return nil
}
