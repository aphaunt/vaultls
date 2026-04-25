package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultls/internal/vault"
)

var cloneOverwrite bool
var cloneDryRun bool

var cloneCmd = &cobra.Command{
	Use:   "clone <src-path> <dst-path>",
	Short: "Clone all secrets from one path to another",
	Args:  cobra.ExactArgs(2),
	RunE:  runClone,
}

func init() {
	cloneCmd.Flags().BoolVar(&cloneOverwrite, "overwrite", false, "Overwrite existing secrets at destination")
	cloneCmd.Flags().BoolVar(&cloneDryRun, "dry-run", false, "Preview changes without writing")
	rootCmd.AddCommand(cloneCmd)
}

func runClone(cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]

	client, err := vault.NewClient(
		cmd.Root().PersistentFlags().Lookup("address").Value.String(),
		cmd.Root().PersistentFlags().Lookup("token").Value.String(),
	)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	opts := vault.CloneOptions{
		Overwrite: cloneOverwrite,
		DryRun:    cloneDryRun,
	}

	result, err := vault.CloneSecrets(context.Background(), client, src, dst, opts)
	if err != nil {
		return err
	}

	prefix := ""
	if cloneDryRun {
		prefix = "[dry-run] "
	}

	for _, k := range result.Copied {
		fmt.Fprintf(cmd.OutOrStdout(), "%scopied: %s\n", prefix, k)
	}
	for _, k := range result.Skipped {
		fmt.Fprintf(cmd.OutOrStdout(), "skipped (exists): %s\n", k)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "done: %d copied, %d skipped\n", len(result.Copied), len(result.Skipped))
	return nil
}
