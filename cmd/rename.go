package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var overwriteRename bool

var renameCmd = &cobra.Command{
	Use:   "rename <source> <destination>",
	Short: "Rename (move) secrets from one path to another",
	Args:  cobra.ExactArgs(2),
	RunE:  runRename,
}

func init() {
	renameCmd.Flags().BoolVar(&overwriteRename, "overwrite", false, "Overwrite destination if it already exists")
	rootCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	src := args[0]
	dst := args[1]

	client, err := vault.NewClient(os.Getenv("VAULT_ADDR"), os.Getenv("VAULT_TOKEN"))
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	summary, err := vault.RenameSecretsWithValidation(cmd.Context(), client, src, dst, overwriteRename)
	if err != nil {
		return err
	}

	if summary.Overwritten {
		fmt.Fprintf(cmd.OutOrStdout(), "Renamed %q → %q (%d key(s) moved, destination overwritten)\n",
			summary.Source, summary.Destination, summary.KeysMoved)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Renamed %q → %q (%d key(s) moved)\n",
			summary.Source, summary.Destination, summary.KeysMoved)
	}
	return nil
}
