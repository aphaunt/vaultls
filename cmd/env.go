package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var envCmd = &cobra.Command{
	Use:   "env <pathA> <pathB>",
	Short: "Compare secrets between two paths as environment variables",
	Long: `Compare secrets stored at two Vault paths and display differences
in an environment-variable-friendly format.

Example:
  vaultls env secret/data/staging secret/data/production`,
	Args: cobra.ExactArgs(2),
	RunE: runEnv,
}

func init() {
	rootCmd.AddCommand(envCmd)
}

func runEnv(cmd *cobra.Command, args []string) error {
	pathA := args[0]
	pathB := args[1]

	if pathA == pathB {
		return fmt.Errorf("pathA and pathB must be different")
	}

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("initializing vault client: %w", err)
	}

	diff, err := vault.CompareEnvs(cmd.Context(), client, pathA, pathB)
	if err != nil {
		return fmt.Errorf("comparing environments: %w", err)
	}

	output := vault.FormatEnvDiff(diff, pathA, pathB)
	fmt.Fprint(os.Stdout, output)

	if len(diff.Changed) > 0 || len(diff.OnlyInA) > 0 || len(diff.OnlyInB) > 0 {
		os.Exit(1)
	}

	return nil
}
