package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var protectCmd = &cobra.Command{
	Use:   "protect <path>",
	Short: "Protect a secret path from accidental overwrites",
	Args:  cobra.ExactArgs(1),
	RunE:  runProtect,
}

var unprotectCmd = &cobra.Command{
	Use:   "unprotect <path>",
	Short: "Remove protection from a secret path",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnprotect,
}

var protectStatusCmd = &cobra.Command{
	Use:   "protect-status <path>",
	Short: "Check whether a secret path is protected",
	Args:  cobra.ExactArgs(1),
	RunE:  runProtectStatus,
}

func init() {
	rootCmd.AddCommand(protectCmd)
	rootCmd.AddCommand(unprotectCmd)
	rootCmd.AddCommand(protectStatusCmd)
}

func runProtect(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}
	if err := vault.ProtectSecrets(cmd.Context(), client, args[0]); err != nil {
		return fmt.Errorf("protect failed: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Protected: %s\n", args[0])
	return nil
}

func runUnprotect(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}
	if err := vault.UnprotectSecrets(cmd.Context(), client, args[0]); err != nil {
		return fmt.Errorf("unprotect failed: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Unprotected: %s\n", args[0])
	return nil
}

func runProtectStatus(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}
	protected, err := vault.IsProtected(cmd.Context(), client, args[0])
	if err != nil {
		return fmt.Errorf("status check failed: %w", err)
	}
	if protected {
		fmt.Fprintf(cmd.OutOrStdout(), "%s is PROTECTED\n", args[0])
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "%s is not protected\n", args[0])
	}
	return nil
}
