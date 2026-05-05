package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var patchDryRun bool

var patchCmd = &cobra.Command{
	Use:   "patch <path> key=value [key=value ...]",
	Short: "Selectively update keys in a Vault secret without overwriting others",
	Example: `  vaultls patch secret/data/myapp DB_PASS=newpass
  vaultls patch secret/data/myapp HOST=db.prod PORT=5432 --dry-run`,
	Args: cobra.MinimumNArgs(2),
	RunE: runPatch,
}

func init() {
	patchCmd.Flags().BoolVar(&patchDryRun, "dry-run", false, "Preview changes without writing to Vault")
	rootCmd.AddCommand(patchCmd)
}

func runPatch(cmd *cobra.Command, args []string) error {
	path := args[0]
	updates := make(map[string]string, len(args)-1)
	for _, pair := range args[1:] {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid key=value pair: %q", pair)
		}
		updates[parts[0]] = parts[1]
	}

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	results, err := vault.PatchSecrets(cmd.Context(), client, path, updates, patchDryRun)
	if err != nil {
		return err
	}

	for _, r := range results {
		if r.Skipped {
			fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] %s: %q -> %q\n", r.Key, r.OldVal, r.NewVal)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "patched %s: %q -> %q\n", r.Key, r.OldVal, r.NewVal)
		}
	}
	return nil
}
