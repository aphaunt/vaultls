package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var (
	flattenSeparator string
	flattenDryRun    bool
)

func init() {
	flattenCmd := &cobra.Command{
		Use:   "flatten <path>",
		Short: "Flatten nested secret keys using a separator",
		Long: `Read all secrets at <path> and collapse any nested structure into
flat keys joined by the separator (default: "__").

With --dry-run the flattened result is printed without writing back to Vault.`,
		Args:    cobra.ExactArgs(1),
		RunE:    runFlatten,
		Example: "  vaultls flatten secret/myapp --separator __",
	}

	flattenCmd.Flags().StringVar(&flattenSeparator, "separator", "__", "Key separator used when joining nested keys")
	flattenCmd.Flags().BoolVar(&flattenDryRun, "dry-run", false, "Print flattened keys without writing back to Vault")

	rootCmd.AddCommand(flattenCmd)
}

func runFlatten(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	result, err := vault.FlattenSecrets(cmd.Context(), client, path, flattenSeparator, flattenDryRun)
	if err != nil {
		return fmt.Errorf("flatten: %w", err)
	}

	if flattenDryRun {
		fmt.Fprintln(os.Stdout, "[dry-run] flattened keys:")
	}

	for _, k := range vault.FlatKeys(result) {
		fmt.Fprintf(os.Stdout, "  %s = %s\n", k, result[k])
	}

	if !flattenDryRun {
		fmt.Fprintf(os.Stdout, "flattened %d key(s) at %s\n", len(result), path)
	}

	return nil
}
