package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var (
	mirrorDeleteOrphans bool
	mirrorDryRun        bool
)

var mirrorCmd = &cobra.Command{
	Use:   "mirror <src> <dst>",
	Short: "Mirror secrets from src to dst, making dst an exact replica",
	Long: `Mirror reads all key/value pairs from src and writes them to dst.
By default, keys in dst that do not exist in src are left untouched.
Use --delete-orphans to remove such keys and achieve a full replica.`,
	Args: cobra.ExactArgs(2),
	RunE: runMirror,
}

func init() {
	rootCmd.AddCommand(mirrorCmd)
	mirrorCmd.Flags().BoolVar(&mirrorDeleteOrphans, "delete-orphans", false, "delete keys in dst that are absent in src")
	mirrorCmd.Flags().BoolVar(&mirrorDryRun, "dry-run", false, "preview changes without writing to Vault")
}

func runMirror(cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("initializing vault client: %w", err)
	}

	result, err := vault.MirrorSecrets(cmd.Context(), client, src, dst, mirrorDeleteOrphans, mirrorDryRun)
	if err != nil {
		return err
	}

	if mirrorDryRun {
		fmt.Fprintln(os.Stdout, "[dry-run] no changes written")
	}

	for _, k := range result.Copied {
		fmt.Fprintf(os.Stdout, "  copied : %s\n", k)
	}
	for _, k := range result.Skipped {
		fmt.Fprintf(os.Stdout, "  skipped: %s (unchanged)\n", k)
	}
	for _, k := range result.Deleted {
		fmt.Fprintf(os.Stdout, "  deleted: %s (orphan)\n", k)
	}

	fmt.Fprintf(os.Stdout, "\nmirror complete: %d copied, %d skipped, %d deleted\n",
		len(result.Copied), len(result.Skipped), len(result.Deleted))
	return nil
}
