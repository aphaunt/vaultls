package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultls/internal/vault"
)

var traceCmd = &cobra.Command{
	Use:   "trace <path>",
	Short: "Trace version history and access metadata for a secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runTrace,
}

func init() {
	rootCmd.AddCommand(traceCmd)
}

func runTrace(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	result, err := vault.TraceSecrets(cmd.Context(), client, path)
	if err != nil {
		return fmt.Errorf("trace failed: %w", err)
	}

	if len(result.Entries) == 0 {
		fmt.Fprintf(os.Stdout, "No trace entries found for %s\n", path)
		return nil
	}

	sort.Slice(result.Entries, func(i, j int) bool {
		return result.Entries[i].Version < result.Entries[j].Version
	})

	fmt.Fprintf(os.Stdout, "Trace for: %s\n", result.Path)
	fmt.Fprintf(os.Stdout, "%-10s %-12s %s\n", "VERSION", "OPERATION", "TIMESTAMP")
	fmt.Fprintf(os.Stdout, "%-10s %-12s %s\n", "-------", "---------", "---------")
	for _, e := range result.Entries {
		fmt.Fprintf(os.Stdout, "%-10d %-12s %s\n", e.Version, e.Operation, e.Caller)
	}
	return nil
}
