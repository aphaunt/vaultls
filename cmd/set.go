package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"vaultls/internal/vault"
)

var setDryRun bool

var setCmd = &cobra.Command{
	Use:   "set <path> key=value [key=value ...]",
	Short: "Set one or more secrets at a Vault path",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runSet,
}

func init() {
	setCmd.Flags().BoolVar(&setDryRun, "dry-run", false, "Preview changes without writing")
	rootCmd.AddCommand(setCmd)
}

func runSet(cmd *cobra.Command, args []string) error {
	path := args[0]
	pairs := map[string]string{}

	for _, arg := range args[1:] {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid key=value pair: %q", arg)
		}
		pairs[parts[0]] = parts[1]
	}

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	results, err := vault.SetSecrets(cmd.Context(), client, path, pairs, setDryRun)
	if err != nil {
		return err
	}

	for _, r := range results {
		switch {
		case r.Written:
			fmt.Fprintf(cmd.OutOrStdout(), "SET   %s\n", r.Key)
		case r.Skipped:
			fmt.Fprintf(cmd.OutOrStdout(), "SKIP  %s (%s)\n", r.Key, r.Reason)
		}
	}
	return nil
}
