package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var mergeOverwrite bool
var mergeDryRun bool

var mergeCmd = &cobra.Command{
	Use:   "merge <src> <dst>",
	Short: "Merge secrets from one path into another",
	Long: `Merge all secrets from <src> into <dst>.

Existing keys in <dst> are skipped unless --overwrite is set.
Use --dry-run to preview changes without writing.`,
	Args: cobra.ExactArgs(2),
	RunE: runMerge,
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().BoolVar(&mergeOverwrite, "overwrite", false, "Overwrite existing keys in dst")
	mergeCmd.Flags().BoolVar(&mergeDryRun, "dry-run", false, "Preview changes without writing")
}

func runMerge(cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]

	client, err := vault.NewClient(
		os.Getenv("VAULT_ADDR"),
		os.Getenv("VAULT_TOKEN"),
	)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	result, err := vault.MergeSecrets(cmd.Context(), client, src, dst, mergeOverwrite, mergeDryRun)
	if err != nil {
		return err
	}

	if mergeDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "[dry-run] No changes written.")
	}

	for _, k := range result.Written {
		fmt.Fprintf(cmd.OutOrStdout(), "  + %s\n", k)
	}
	for _, k := range result.Overwritten {
		fmt.Fprintf(cmd.OutOrStdout(), "  ~ %s\n", k)
	}
	for _, k := range result.Skipped {
		fmt.Fprintf(cmd.OutOrStdout(), "  = %s (skipped)\n", k)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nMerge complete: %d written, %d overwritten, %d skipped.\n",
		len(result.Written), len(result.Overwritten), len(result.Skipped))
	return nil
}
