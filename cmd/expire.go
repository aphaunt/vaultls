package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var expireCmd = &cobra.Command{
	Use:   "expire <path>",
	Short: "Show expiry status of secrets at a path",
	Long: `Reads custom_metadata from Vault KV v2 and reports which secrets
have an 'expires_at' timestamp set, whether they are expired, and their TTL.`,
	Args:    cobra.ExactArgs(1),
	RunE:    runExpire,
}

var expireOnlyExpired bool

func init() {
	rootCmd.AddCommand(expireCmd)
	expireCmd.Flags().BoolVar(&expireOnlyExpired, "expired", false, "Show only expired secrets")
}

func runExpire(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	now := time.Now()
	results, err := vault.ExpireSecrets(cmd.Context(), client, path, now)
	if err != nil {
		return fmt.Errorf("expire check failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No expiry metadata found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tEXPIRES AT\tSTATUS\tTTL")
	for _, r := range results {
		if expireOnlyExpired && !r.Expired {
			continue
		}
		status := "valid"
		ttlStr := r.TTL.Round(time.Second).String()
		if r.Expired {
			status = "EXPIRED"
			ttlStr = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			r.Key,
			r.ExpiresAt.Format(time.RFC3339),
			status,
			ttlStr,
		)
	}
	return w.Flush()
}
