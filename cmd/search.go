package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultls/internal/vault"
)

var searchValueQuery string

var searchCmd = &cobra.Command{
	Use:   "search <path>",
	Short: "Search secrets by key or value substring",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().StringVarP(&searchValueQuery, "value", "v", "", "Substring to search for in secret values")
	searchCmd.Flags().StringVarP(&searchQuery, "key", "k", "", "Substring to search for in secret keys")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient(
		os.Getenv("VAULT_ADDR"),
		os.Getenv("VAULT_TOKEN"),
	)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	ctx := context.Background()

	if searchValueQuery != "" {
		results, err := vault.SearchSecretsByValue(ctx, client, path, searchValueQuery)
		if err != nil {
			return err
		}
		if len(results) == 0 {
			fmt.Println("No matches found.")
			return nil
		}
		for _, r := range results {
			fmt.Printf("[%s]\n", r.Path)
			for k, v := range r.Matches {
				fmt.Printf("  %s = %s\n", k, v)
			}
		}
		return nil
	}

	// Fall back to key search
	matches, err := vault.SearchSecrets(ctx, client, path, searchQuery)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		fmt.Println("No matches found.")
		return nil
	}
	for _, m := range matches {
		fmt.Println(m)
	}
	return nil
}
