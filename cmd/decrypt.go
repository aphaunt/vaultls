package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var (
	decryptTransitKey string
	decryptPatterns   []string
	decryptDryRun     bool
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt <path>",
	Short: "Decrypt secret values at a Vault path using the Transit engine",
	Long: `Reads secrets at the given KV path, decrypts values whose keys match
one or more regex patterns using Vault's Transit secrets engine, and writes
the plaintext values back unless --dry-run is set.`,
	Args: cobra.ExactArgs(1),
	RunE: runDecrypt,
}

func init() {
	rootCmd.AddCommand(decryptCmd)
	decryptCmd.Flags().StringVar(&decryptTransitKey, "transit-key", "", "Transit key name used for decryption (required)")
	decryptCmd.Flags().StringArrayVar(&decryptPatterns, "pattern", nil, "Regex pattern(s) to match secret keys for decryption (required)")
	decryptCmd.Flags().BoolVar(&decryptDryRun, "dry-run", false, "Print changes without writing back to Vault")
	_ = decryptCmd.MarkFlagRequired("transit-key")
	_ = decryptCmd.MarkFlagRequired("pattern")
}

func runDecrypt(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	results, err := vault.DecryptSecrets(cmd.Context(), client, path, decryptTransitKey, decryptPatterns, decryptDryRun)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	if decryptDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "[dry-run] would write decrypted values:")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Decrypted values written:")
	}

	for k, v := range results {
		matched := false
		for _, p := range decryptPatterns {
			if strings.Contains(k, p) {
				matched = true
				break
			}
		}
		if matched {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s = %s\n", k, v)
		}
	}

	return nil
}
