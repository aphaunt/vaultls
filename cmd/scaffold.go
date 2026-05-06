package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"vaultls/internal/vault"
)

var (
	scaffoldKeys      []string
	scaffoldDefaults  []string
	scaffoldDryRun    bool
	scaffoldOverwrite bool
)

func init() {
	scaffoldCmd := &cobra.Command{
		Use:   "scaffold <path>",
		Short: "Write a skeleton secret with placeholder keys",
		Args:  cobra.ExactArgs(1),
		RunE:  runScaffold,
	}
	scaffoldCmd.Flags().StringSliceVarP(&scaffoldKeys, "keys", "k", nil, "Comma-separated list of keys to scaffold (required)")
	scaffoldCmd.Flags().StringSliceVarP(&scaffoldDefaults, "defaults", "d", nil, "Default values as key=value pairs")
	scaffoldCmd.Flags().BoolVar(&scaffoldDryRun, "dry-run", false, "Preview without writing")
	scaffoldCmd.Flags().BoolVar(&scaffoldOverwrite, "overwrite", false, "Overwrite existing secret")
	_ = scaffoldCmd.MarkFlagRequired("keys")
	rootCmd.AddCommand(scaffoldCmd)
}

func runScaffold(cmd *cobra.Command, args []string) error {
	path := args[0]

	defaultsMap := make(map[string]string)
	for _, pair := range scaffoldDefaults {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid default pair %q: expected key=value", pair)
		}
		defaultsMap[parts[0]] = parts[1]
	}

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	result, err := vault.ScaffoldSecrets(cmd.Context(), client, vault.ScaffoldOptions{
		Path:      path,
		Keys:      scaffoldKeys,
		Defaults:  defaultsMap,
		DryRun:    scaffoldDryRun,
		Overwrite: scaffoldOverwrite,
	})
	if err != nil {
		return err
	}

	if scaffoldDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "[dry-run] would write:")
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "scaffolded %d key(s) at %s\n", len(result), path)
	}
	for k, v := range result {
		if v == "" {
			v = "(empty)"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s = %s\n", k, v)
	}
	return nil
}
