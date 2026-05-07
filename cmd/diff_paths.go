package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var diffPathsCmd = &cobra.Command{
	Use:   "diff-paths <pathA> <pathB>",
	Short: "Compare keys present at two vault paths",
	Args:  cobra.ExactArgs(2),
	RunE:  runDiffPaths,
}

func init() {
	rootCmd.AddCommand(diffPathsCmd)
}

func runDiffPaths(cmd *cobra.Command, args []string) error {
	pathA := args[0]
	pathB := args[1]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	res, err := vault.DiffPaths(cmd.Context(), client, pathA, pathB)
	if err != nil {
		return fmt.Errorf("diff-paths: %w", err)
	}

	w := os.Stdout

	if len(res.OnlyInA) == 0 && len(res.OnlyInB) == 0 {
		fmt.Fprintln(w, "Paths are identical (same key set).")
		return nil
	}

	sort.Strings(res.OnlyInA)
	sort.Strings(res.OnlyInB)
	sort.Strings(res.InBoth)

	if len(res.OnlyInA) > 0 {
		fmt.Fprintf(w, "Only in %s:\n", pathA)
		for _, k := range res.OnlyInA {
			fmt.Fprintf(w, "  - %s\n", k)
		}
	}

	if len(res.OnlyInB) > 0 {
		fmt.Fprintf(w, "Only in %s:\n", pathB)
		for _, k := range res.OnlyInB {
			fmt.Fprintf(w, "  + %s\n", k)
		}
	}

	if len(res.InBoth) > 0 {
		fmt.Fprintf(w, "In both:\n")
		for _, k := range res.InBoth {
			fmt.Fprintf(w, "    %s\n", k)
		}
	}

	return nil
}
