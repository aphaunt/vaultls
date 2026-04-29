package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/user/vaultls/internal/vault"
)

var snapshotOutput string

var snapshotCmd = &cobra.Command{
	Use:   "snapshot <path>",
	Short: "Take a point-in-time snapshot of secrets at a path",
	Args:  cobra.ExactArgs(1),
	RunE:  runSnapshot,
}

func init() {
	snapshotCmd.Flags().StringVarP(&snapshotOutput, "output", "o", "", "Write snapshot JSON to file instead of stdout")
	rootCmd.AddCommand(snapshotCmd)
}

func runSnapshot(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	snap, err := vault.TakeSnapshot(cmd.Context(), client, path)
	if err != nil {
		return fmt.Errorf("taking snapshot: %w", err)
	}

	data, err := vault.SnapshotToJSON(snap)
	if err != nil {
		return fmt.Errorf("serializing snapshot: %w", err)
	}

	if snapshotOutput != "" {
		if err := os.WriteFile(snapshotOutput, data, 0o600); err != nil {
			return fmt.Errorf("writing snapshot to %q: %w", snapshotOutput, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Snapshot written to %s\n", snapshotOutput)
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}
