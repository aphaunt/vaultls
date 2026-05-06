package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"vaultls/internal/vault"
)

var splitDryRun bool
var splitDefault string

var splitCmd = &cobra.Command{
	Use:   "split <src-path> <key=dest>...",
	Short: "Split secrets from one path into multiple destination paths",
	Long: `Reads secrets from a source KV path and distributes individual keys
to separate destination paths based on a key=destination mapping.

Example:
  vaultls split secret/app DB_HOST=secret/db DB_PASS=secret/db API_KEY=secret/api`,
	Args: cobra.MinimumNArgs(2),
	RunE: runSplit,
}

func init() {
	rootCmd.AddCommand(splitCmd)
	splitCmd.Flags().BoolVar(&splitDryRun, "dry-run", false, "Preview split without writing to Vault")
	splitCmd.Flags().StringVar(&splitDefault, "default-dest", "", "Destination path for keys not in the mapping")
}

func runSplit(cmd *cobra.Command, args []string) error {
	srcPath := args[0]
	mapping := map[string]string{}

	for _, pair := range args[1:] {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid mapping %q: expected format key=destination", pair)
		}
		mapping[parts[0]] = parts[1]
	}

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	result, err := vault.SplitSecrets(cmd.Context(), client, srcPath, mapping, splitDefault, splitDryRun)
	if err != nil {
		return err
	}

	if splitDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "[dry-run] Split preview:")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Split complete:")
	}

	dests := make([]string, 0, len(result))
	for d := range result {
		dests = append(dests, d)
	}
	sort.Strings(dests)

	for _, dest := range dests {
		keys := result[dest]
		sort.Strings(keys)
		fmt.Fprintf(cmd.OutOrStdout(), "  %s <- [%s]\n", dest, strings.Join(keys, ", "))
	}

	return nil
}
