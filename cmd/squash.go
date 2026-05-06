package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var (
	squashDest      string
	squashDryRun    bool
	squashOverwrite bool
	squashPrefix    bool
)

var squashCmd = &cobra.Command{
	Use:   "squash <src1> [src2 ...] --dest <dest>",
	Short: "Merge multiple secret paths into a single destination path",
	Long: `Squash reads secrets from one or more KV paths and writes them
all into a single destination path. Keys from later sources overwrite
earlier ones when --overwrite is set. Use --prefix to namespace keys
with the last segment of their source path.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSquash,
}

func init() {
	rootCmd.AddCommand(squashCmd)
	squashCmd.Flags().StringVar(&squashDest, "dest", "", "Destination KV path (required)")
	squashCmd.Flags().BoolVar(&squashDryRun, "dry-run", false, "Preview without writing")
	squashCmd.Flags().BoolVar(&squashOverwrite, "overwrite", false, "Overwrite duplicate keys with later source values")
	squashCmd.Flags().BoolVar(&squashPrefix, "prefix", false, "Prefix keys with their source path segment")
	_ = squashCmd.MarkFlagRequired("dest")
}

func runSquash(cmd *cobra.Command, args []string) error {
	if squashDest == "" {
		return fmt.Errorf("--dest flag is required")
	}

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	opts := vault.SquashOptions{
		Paths:     args,
		Dest:      squashDest,
		DryRun:    squashDryRun,
		Overwrite: squashOverwrite,
		Prefix:    squashPrefix,
	}

	result, err := vault.SquashSecrets(cmd.Context(), client, opts)
	if err != nil {
		return err
	}

	if squashDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "[dry-run] would write to:", squashDest)
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "squashed to:", squashDest)
	}

	keys := make([]string, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "keys:", strings.Join(keys, ", "))
	return nil
}
