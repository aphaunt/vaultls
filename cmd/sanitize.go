package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var sanitizeDryRun bool

var sanitizeCmd = &cobra.Command{
	Use:   "sanitize <path>",
	Short: "Trim whitespace from secret values at a given path",
	Long: `Reads all string values at the specified Vault path and trims
leading/trailing whitespace. Use --dry-run to preview changes without writing.`,
	Args: cobra.ExactArgs(1),
	RunE: runSanitize,
}

func init() {
	sanitizeCmd.Flags().BoolVar(&sanitizeDryRun, "dry-run", false, "Preview changes without writing to Vault")
	rootCmd.AddCommand(sanitizeCmd)
}

func runSanitize(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}

	results, err := vault.SanitizeSecrets(cmd.Context(), client, path, sanitizeDryRun)
	if err != nil {
		return fmt.Errorf("sanitize failed: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tOLD\tNEW\tCHANGED")

	changedCount := 0
	for _, r := range results {
		mark := ""
		if r.Changed {
			mark = "yes"
			changedCount++
		}
		fmt.Fprintf(w, "%s\t%q\t%q\t%s\n", r.Key, r.OldValue, r.NewValue, mark)
	}
	w.Flush()

	if sanitizeDryRun {
		fmt.Fprintf(os.Stdout, "\n[dry-run] %d key(s) would be updated.\n", changedCount)
	} else {
		fmt.Fprintf(os.Stdout, "\n%d key(s) updated.\n", changedCount)
	}
	return nil
}
