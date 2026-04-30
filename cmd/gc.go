package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var gcDryRun bool
var gcPatterns []string

var gcCmd = &cobra.Command{
	Use:   "gc <path>",
	Short: "Remove secrets whose keys match given patterns",
	Long: `Garbage-collect secrets at a KV path by deleting keys that match
one or more substring patterns. Use --dry-run to preview removals.`,
	Args: cobra.ExactArgs(1),
	RunE: runGC,
}

func init() {
	gcCmd.Flags().BoolVar(&gcDryRun, "dry-run", false, "Preview keys that would be deleted without making changes")
	gcCmd.Flags().StringSliceVar(&gcPatterns, "pattern", nil, "Substring patterns to match against key names (repeatable)")
	_ = gcCmd.MarkFlagRequired("pattern")
	rootCmd.AddCommand(gcCmd)
}

func runGC(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("initializing vault client: %w", err)
	}

	results, err := vault.GCSecrets(cmd.Context(), client, path, gcPatterns, gcDryRun)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No matching keys found.")
		return nil
	}

	prefix := "deleted"
	if gcDryRun {
		prefix = "would delete"
	}

	for _, r := range results {
		short := strings.TrimPrefix(r.Path, path+"/")
		fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", prefix, short)
	}

	if gcDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "(dry-run: no changes written)")
	}
	return nil
}
