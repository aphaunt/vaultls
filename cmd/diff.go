package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"vaultls/internal/vault"
)

var (
	envAAddr  string
	envAToken string
	envBAddr  string
	envBToken string
	noColor   bool
)

var diffCmd = &cobra.Command{
	Use:   "diff <path>",
	Short: "Diff secrets at a path between two Vault environments",
	Args:  cobra.ExactArgs(1),
	RunE:  runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().StringVar(&envAAddr, "addr-a", "", "Vault address for environment A (overrides VAULT_ADDR)")
	diffCmd.Flags().StringVar(&envAToken, "token-a", "", "Vault token for environment A (overrides VAULT_TOKEN)")
	diffCmd.Flags().StringVar(&envBAddr, "addr-b", "", "Vault address for environment B")
	diffCmd.Flags().StringVar(&envBToken, "token-b", "", "Vault token for environment B")
	diffCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	_ = diffCmd.MarkFlagRequired("addr-b")
	_ = diffCmd.MarkFlagRequired("token-b")
}

func runDiff(cmd *cobra.Command, args []string) error {
	path := args[0]

	addrA := envAAddr
	if addrA == "" {
		addrA = viper.GetString("address")
	}
	tokenA := envAToken
	if tokenA == "" {
		tokenA = viper.GetString("token")
	}

	clientA, err := vault.NewClient(addrA, tokenA)
	if err != nil {
		return fmt.Errorf("env A client: %w", err)
	}
	clientB, err := vault.NewClient(envBAddr, envBToken)
	if err != nil {
		return fmt.Errorf("env B client: %w", err)
	}

	secretsA, err := clientA.GetSecrets(cmd.Context(), path)
	if err != nil {
		return fmt.Errorf("reading env A secrets: %w", err)
	}
	secretsB, err := clientB.GetSecrets(cmd.Context(), path)
	if err != nil {
		return fmt.Errorf("reading env B secrets: %w", err)
	}

	result := vault.DiffSecrets(secretsA, secretsB)
	vault.RenderDiff(os.Stdout, result, addrA, envBAddr, !noColor)
	return nil
}
