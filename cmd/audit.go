package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/vaultls/internal/vault"
)

var auditCmd = &cobra.Command{
	Use:   "audit <path>",
	Short: "Show recent audit log entries for a Vault path",
	Args:  cobra.ExactArgs(1),
	RunE:  runAudit,
}

var auditOperation string

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.Flags().StringVarP(&auditOperation, "operation", "o", "", "Filter by operation type (read, write, etc.)")
}

func runAudit(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	log, err := vault.GetAuditLog(cmd.Context(), client, path)
	if err != nil {
		return fmt.Errorf("fetching audit log for %q: %w", path, err)
	}

	entries := log.Entries
	if auditOperation != "" {
		entries = log.FilterByOperation(auditOperation)
	}

	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No audit entries found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIME\tTYPE\tOPERATION\tPATH")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Time.Format("2006-01-02 15:04:05"), e.Type, e.Operation, e.Path)
	}
	return w.Flush()
}
