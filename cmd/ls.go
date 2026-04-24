package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/user/vaultls/internal/vault"
)

var lsCmd = &cobra.Command{
	Use:   "ls [path]",
	Short: "List secrets at a given Vault path",
	Args:  cobra.ExactArgs(1),
	RunE:  runLS,
}

func init() {
	rootCmd.AddCommand(lsCmd)
}

func runLS(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient(vault.Config{
		Address: viper.GetString("vault.addr"),
		Token:   viper.GetString("vault.token"),
	})
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	keys, err := client.List(path)
	if err != nil {
		return fmt.Errorf("failed to list path: %w", err)
	}

	if len(keys) == 0 {
		fmt.Fprintf(os.Stderr, "No secrets found at path: %s\n", path)
		return nil
	}

	for _, key := range keys {
		prefix := "  "
		if strings.HasSuffix(key, "/") {
			prefix = "📁"
		} else {
			prefix = "🔑"
		}
		fmt.Printf("%s %s\n", prefix, key)
	}

	return nil
}
