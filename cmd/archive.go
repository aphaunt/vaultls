package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var archiveCmd = &cobra.Command{
	Use:   "archive <src> <archive-root>",
	Short: "Archive secrets from a path to an archive root",
	Long: `Reads all secrets under <src> and writes them to <archive-root>.
Optionally deletes the originals after archiving.`,
	Args: cobra.ExactArgs(2),
	RunE: runArchive,
}

var (
	archiveDelete bool
	archiveDryRun bool
)

func init() {
	archiveCmd.Flags().BoolVar(&archiveDelete, "delete", false, "Delete originals after archiving")
	archiveCmd.Flags().BoolVar(&archiveDryRun, "dry-run", false, "Preview archive actions without writing")
	rootCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) error {
	src := args[0]
	archiveRoot := args[1]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	results, err := vault.ArchiveSecrets(cmd.Context(), client, src, archiveRoot, archiveDelete, archiveDryRun)
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tSTATUS\tREASON")
	for _, r := range results {
		status := "archived"
		if r.Skipped {
			status = "skipped"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", r.Path, status, r.Reason)
	}
	w.Flush()

	return nil
}
