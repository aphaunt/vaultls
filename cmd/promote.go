package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"vaultls/internal/vault"
)

var promoteOverwrite bool

var promoteCmd = &cobra.Command{
	Use:   "promote <src-path> <dst-path>",
	Short: "Promote secrets from one path to another (e.g. dev → prod)",
	Args:  cobra.ExactArgs(2),
	RunE:  runPromote,
}

func init() {
	promoteCmd.Flags().BoolVar(&promoteOverwrite, "overwrite", false, "Overwrite existing keys at destination")
	rootCmd.AddCommand(promoteCmd)
}

func runPromote(cmd *cobra.Command, args []string) error {
	srcPath := args[0]
	dstPath := args[1]

	client, err := vault.NewClient(
		cmd.Root().PersistentFlags().Lookup("address").Value.String(),
		cmd.Root().PersistentFlags().Lookup("token").Value.String(),
	)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	result, err := vault.PromoteSecrets(cmd.Context(), client, srcPath, dstPath, promoteOverwrite)
	if err != nil {
		return err
	}

	w := os.Stdout
	fmt.Fprintf(w, "Promote: %s → %s\n", srcPath, dstPath)
	for _, k := range result.Copied {
		fmt.Fprintf(w, "  + %s (copied)\n", k)
	}
	for _, k := range result.Overwritten {
		fmt.Fprintf(w, "  ~ %s (overwritten)\n", k)
	}
	for _, k := range result.Skipped {
		fmt.Fprintf(w, "  - %s (skipped, already exists)\n", k)
	}
	fmt.Fprintf(w, "Done. copied=%d overwritten=%d skipped=%d\n",
		len(result.Copied), len(result.Overwritten), len(result.Skipped))
	return nil
}
