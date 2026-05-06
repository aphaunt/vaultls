package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var (
	encryptTransitKey string
	encryptPatterns   []string
	encryptDryRun     bool
)

func init() {
	encryptCmd := &cobra.Command{
		Use:   "encrypt <path>",
		Short: "Encrypt secret values in-place using Vault Transit",
		Long: `Reads secrets at <path>, encrypts values whose keys match the given
patterns using the specified Vault Transit key, and writes them back.

Example:
  vaultls encrypt secret/data/myapp --transit-key mykey --patterns DB_PASS,API_KEY`,
		Args: cobra.ExactArgs(1),
		RunE: runEncrypt,
	}

	encryptCmd.Flags().StringVar(&encryptTransitKey, "transit-key", "", "Vault Transit key name to use for encryption (required)")
	encryptCmd.Flags().StringSliceVar(&encryptPatterns, "patterns", nil, "Comma-separated list of key patterns to encrypt (required)")
	encryptCmd.Flags().BoolVar(&encryptDryRun, "dry-run", false, "Preview changes without writing to Vault")
	_ = encryptCmd.MarkFlagRequired("transit-key")
	_ = encryptCmd.MarkFlagRequired("patterns")

	rootCmd.AddCommand(encryptCmd)
}

func runEncrypt(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("initializing vault client: %w", err)
	}

	results, err := vault.EncryptSecrets(cmd.Context(), client, path, encryptTransitKey, encryptPatterns, encryptDryRun)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No matching keys found to encrypt.")
		return nil
	}

	if encryptDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "[dry-run] Keys that would be encrypted:")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Encrypted keys:")
	}

	for _, r := range results {
		masked := strings.Repeat("*", len(r.Original))
		fmt.Fprintf(cmd.OutOrStdout(), "  %-30s %s -> %s\n", r.Key, masked, r.Cipher)
	}
	return nil
}
