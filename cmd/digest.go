package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var digestCompare bool

var digestCmd = &cobra.Command{
	Use:   "digest <path> [path2]",
	Short: "Compute or compare SHA-256 digests of Vault secret paths",
	Long: `Compute a stable SHA-256 digest over all key-value pairs at a Vault path.
When two paths are provided with --compare, prints whether they match.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runDigest,
}

func init() {
	rootCmd.AddCommand(digestCmd)
	digestCmd.Flags().BoolVar(&digestCompare, "compare", false, "compare digests of two paths")
}

func runDigest(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("vault client error: %w", err)
	}

	ctx := context.Background()

	if digestCompare {
		if len(args) != 2 {
			return fmt.Errorf("--compare requires exactly two paths")
		}
		match, err := vault.CompareDigests(ctx, client, args[0], args[1])
		if err != nil {
			return err
		}
		if match {
			fmt.Fprintln(cmd.OutOrStdout(), "MATCH: digests are identical")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "DIFFER: digests do not match")
		}
		return nil
	}

	result, err := vault.DigestSecrets(ctx, client, args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "path:   %s\ndigest: %s\nkeys:   %d\n",
		result.Path, result.Digest, result.Keys)
	return nil
}
