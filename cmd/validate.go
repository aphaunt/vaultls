package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"vaultls/internal/vault"
)

var validateCmd = &cobra.Command{
	Use:   "validate <path> <key1,key2,...>",
	Short: "Validate that required keys exist and are non-empty at a secret path",
	Args:  cobra.ExactArgs(2),
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	path := args[0]
	keys := strings.Split(args[1], ",")

	for i, k := range keys {
		keys[i] = strings.TrimSpace(k)
	}

	client, err := vault.NewClient(
		cmd.Root().PersistentFlags().Lookup("address").Value.String(),
		cmd.Root().PersistentFlags().Lookup("token").Value.String(),
	)
	if err != nil {
		return fmt.Errorf("client error: %w", err)
	}

	result, err := vault.ValidateSecrets(context.Background(), client, path, keys)
	if err != nil {
		return err
	}

	if len(result.Valid) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "✔ valid:   %s\n", strings.Join(result.Valid, ", "))
	}
	if len(result.Empty) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "⚠ empty:   %s\n", strings.Join(result.Empty, ", "))
	}
	if len(result.Missing) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "✘ missing: %s\n", strings.Join(result.Missing, ", "))
	}

	if !result.IsValid() {
		return fmt.Errorf("validation failed for path %q", path)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "validation passed")
	return nil
}
