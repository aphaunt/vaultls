package cmd

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var (
	rotateDryRun bool
	rotateLength int
)

var rotateCmd = &cobra.Command{
	Use:   "rotate <path>",
	Short: "Rotate secret values at a given Vault path",
	Long: `Re-generates every value at the specified KV-v2 path.
By default a random alphanumeric string of --length characters is used.
Pass --dry-run to preview changes without writing to Vault.`,
	Args: cobra.ExactArgs(1),
	RunE: runRotate,
}

func init() {
	rotateCmd.Flags().BoolVar(&rotateDryRun, "dry-run", false, "Preview changes without writing to Vault")
	rotateCmd.Flags().IntVar(&rotateLength, "length", 32, "Length of generated secret values")
	rootCmd.AddCommand(rotateCmd)
}

func runRotate(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	generator := func(key, _ string) (string, error) {
		return randomString(rotateLength), nil
	}

	results, err := vault.RotateSecrets(cmd.Context(), client, path, generator, rotateDryRun)
	if err != nil {
		return err
	}

	for _, r := range results {
		if rotateDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would rotate %s: %s → %s\n", r.Key, r.OldValue, r.NewValue)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "rotated %s\n", r.Key)
		}
	}

	if rotateDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "dry-run complete, no changes written")
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "rotated %d key(s) at %s\n", len(results), path)
	}
	return nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(n int) string {
	var sb strings.Builder
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}
