package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var (
	tagOverwrite bool
	tagList      bool
)

var tagCmd = &cobra.Command{
	Use:   "tag <path> [key=value ...]",
	Short: "Tag a secret with custom metadata key=value pairs",
	Long: `Attach or update custom metadata tags on a KV v2 secret.

Examples:
  vaultls tag secret/myapp/config env=prod team=platform
  vaultls tag --list secret/myapp/config`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("path is required")
		}
		if !tagList && len(args) < 2 {
			return fmt.Errorf("at least one key=value tag is required")
		}
		return nil
	},
	RunE: runTag,
}

func init() {
	rootCmd.AddCommand(tagCmd)
	tagCmd.Flags().BoolVar(&tagOverwrite, "overwrite", false, "Overwrite existing tag values")
	tagCmd.Flags().BoolVarP(&tagList, "list", "l", false, "List existing tags instead of setting them")
}

func runTag(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	ctx := cmd.Context()

	if tagList {
		tags, err := vault.ListTags(ctx, client, path)
		if err != nil {
			return fmt.Errorf("list tags: %w", err)
		}
		if len(tags) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "(no tags)")
			return nil
		}
		for k, v := range tags {
			fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", k, v)
		}
		return nil
	}

	tags := map[string]string{}
	for _, pair := range args[1:] {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid tag format %q: expected key=value", pair)
		}
		tags[parts[0]] = parts[1]
	}

	if err := vault.TagSecrets(ctx, client, path, tags, tagOverwrite); err != nil {
		return fmt.Errorf("tag: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Tagged %s with %d key(s)\n", path, len(tags))
	return nil
}
