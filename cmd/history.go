package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/user/vaultls/internal/vault"
)

var (
	historyVersionA int
	historyVersionB int
)

var historyCmd = &cobra.Command{
	Use:   "history <path>",
	Short: "Diff two versions of the same secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().IntVarP(&historyVersionA, "version-a", "a", 0, "First version to compare (required)")
	historyCmd.Flags().IntVarP(&historyVersionB, "version-b", "b", 0, "Second version to compare (required)")
	_ = historyCmd.MarkFlagRequired("version-a")
	_ = historyCmd.MarkFlagRequired("version-b")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	path := args[0]

	if historyVersionA == historyVersionB {
		return fmt.Errorf("version-a and version-b must differ (both are %s)",
			strconv.Itoa(historyVersionA))
	}

	client, err := vault.NewClient(
		os.Getenv("VAULT_ADDR"),
		os.Getenv("VAULT_TOKEN"),
		os.Getenv("VAULT_MOUNT"),
	)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	ctx := context.Background()
	diffs, err := client.DiffVersions(ctx, path, historyVersionA, historyVersionB)
	if err != nil {
		return fmt.Errorf("diffing versions: %w", err)
	}

	label := fmt.Sprintf("%s (v%d vs v%d)", path, historyVersionA, historyVersionB)
	vault.RenderDiff(cmd.OutOrStdout(), label, diffs)
	return nil
}
