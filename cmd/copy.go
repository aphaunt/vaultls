package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/vaultls/internal/vault"
)

var copyOverwrite bool

var copyCmd = &cobra.Command{
	Use:   "copy <src-path> <dst-path>",
	Short: "Copy secrets from one Vault path to another",
	Args:  cobra.ExactArgs(2),
	RunE:  runCopy,
}

func init() {
	copyCmd.Flags().BoolVar(&copyOverwrite, "overwrite", false, "Overwrite existing keys at the destination")
	rootCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
	srcPath := args[0]
	dstPath := args[1]

	if srcPath == dstPath {
		return fmt.Errorf("source and destination paths must differ")
	}

	client, err := vault.NewClient(
		cmd.Root().PersistentFlags().Lookup("address").Value.String(),
		cmd.Root().PersistentFlags().Lookup("token").Value.String(),
	)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	result, err := vault.CopySecrets(context.Background(), client, srcPath, dstPath, copyOverwrite)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Copied %d key(s) from %q to %q\n",
		result.KeysCopied, result.Source, result.Destination)
	return nil
}
