package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"vaultls/internal/vault"
)

var injectDryRun bool

var injectCmd = &cobra.Command{
	Use:   "inject <path>",
	Short: "Inject Vault secrets into the current environment",
	Long: `Read secrets from a Vault KV path and inject them as environment
variables into the current process. Use --dry-run to preview without
modifying the environment.`,
	Args: cobra.ExactArgs(1),
	RunE: runInject,
}

func init() {
	injectCmd.Flags().BoolVar(&injectDryRun, "dry-run", false, "Preview variables without setting them")
	rootCmd.AddCommand(injectCmd)
}

func runInject(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}

	injected, err := vault.InjectSecrets(cmd.Context(), client, path, injectDryRun)
	if err != nil {
		return err
	}

	if len(injected) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No secrets found.")
		return nil
	}

	if injectDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "[dry-run] Would inject:")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Injected:")
	}

	for _, entry := range injected {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", parts[0])
		}
	}

	return nil
}
