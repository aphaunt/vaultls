package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultls/internal/vault"
)

var (
	watchInterval   time.Duration
	watchMaxChanges int
)

var watchCmd = &cobra.Command{
	Use:   "watch <path>",
	Short: "Watch a Vault secret path for changes",
	Long: `Continuously polls a Vault KV path and prints any detected changes.

Press Ctrl+C to stop watching.`,
	Args: cobra.ExactArgs(1),
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().DurationVarP(&watchInterval, "interval", "i", 5*time.Second,
		"Polling interval (e.g. 5s, 1m)")
	watchCmd.Flags().IntVarP(&watchMaxChanges, "max-changes", "n", 0,
		"Stop after this many change events (0 = unlimited)")
}

func runWatch(cmd *cobra.Command, args []string) error {
	path := args[0]

	address := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")

	client, err := vault.NewClient(address, token)
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Watching %s every %s...\n", path, watchInterval)

	opts := vault.WatchOptions{
		Interval:   watchInterval,
		MaxChanges: watchMaxChanges,
		Out:        cmd.OutOrStdout(),
	}

	return vault.WatchSecretsWithOutput(cmd.Context(), client, path, opts)
}
