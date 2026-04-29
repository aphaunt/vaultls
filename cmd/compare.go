package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var compareCmd = &cobra.Command{
	Use:   "compare <pathA> <pathB>",
	Short: "Compare secrets between two vault paths",
	Args:  cobra.ExactArgs(2),
	RunE:  runCompare,
}

func init() {
	rootCmd.AddCommand(compareCmd)
}

func runCompare(cmd *cobra.Command, args []string) error {
	pathA := args[0]
	pathB := args[1]

	if pathA == pathB {
		return fmt.Errorf("paths must be different")
	}

	client, err := vault.NewClient(
		os.Getenv("VAULT_ADDR"),
		os.Getenv("VAULT_TOKEN"),
	)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	result, err := vault.CompareSecrets(cmd.Context(), client, pathA, pathB)
	if err != nil {
		return fmt.Errorf("comparing secrets: %w", err)
	}

	if len(result.OnlyInA) == 0 && len(result.OnlyInB) == 0 && len(result.Changed) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No differences found.")
		return nil
	}

	w := cmd.OutOrStdout()
	for _, k := range result.OnlyInA {
		fmt.Fprintf(w, "< %s (only in %s)\n", k, pathA)
	}
	for _, k := range result.OnlyInB {
		fmt.Fprintf(w, "> %s (only in %s)\n", k, pathB)
	}
	for k, pair := range result.Changed {
		fmt.Fprintf(w, "~ %s: %q -> %q\n", k, pair[0], pair[1])
	}

	return nil
}
