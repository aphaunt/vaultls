package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var (
	grepKeysOnly bool
)

var grepCmd = &cobra.Command{
	Use:   "grep <path> <pattern>",
	Short: "Search secrets by key or value pattern",
	Long: `Search all secrets under a Vault path for keys or values matching
a regular expression pattern. Use --keys-only to restrict matching to key names.`,
	Args:    cobra.ExactArgs(2),
	RunE:    runGrep,
	Example: `  vaultls grep secret/myapp "DB_.*"\n  vaultls grep secret/myapp "localhost" --keys-only`,
}

func init() {
	grepCmd.Flags().BoolVar(&grepKeysOnly, "keys-only", false, "Match against key names only, not values")
	rootCmd.AddCommand(grepCmd)
}

func runGrep(cmd *cobra.Command, args []string) error {
	path := args[0]
	pattern := args[1]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("initializing vault client: %w", err)
	}

	results, err := vault.GrepSecrets(context.Background(), client, path, pattern, grepKeysOnly)
	if err != nil {
		return fmt.Errorf("grep failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No matches found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tKEY\tVALUE")
	for _, r := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\n", r.Path, r.Key, r.Value)
	}
	return w.Flush()
}
