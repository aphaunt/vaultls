package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"vaultls/internal/vault"
)

var annotateCmd = &cobra.Command{
	Use:   "annotate <path> <key=value>...",
	Short: "Add or update metadata annotations on a secret",
	Long: `Annotate stores key=value pairs as custom_metadata annotations
on a KV v2 secret path. Annotations are prefixed with "annotation/"
internally to avoid collisions with other metadata fields.`,
	Args: cobra.MinimumNArgs(2),
	RunE: runAnnotate,
}

var (
	annotateDryRun bool
	annotateGet    bool
)

func init() {
	root := rootCmd
	root.AddCommand(annotateCmd)
	annotateCmd.Flags().BoolVar(&annotateDryRun, "dry-run", false, "Preview annotations without writing")
	annotateCmd.Flags().BoolVar(&annotateGet, "get", false, "Read and display existing annotations")
}

func runAnnotate(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	if annotateGet {
		anns, err := vault.GetAnnotations(cmd.Context(), client, path)
		if err != nil {
			return fmt.Errorf("failed to get annotations: %w", err)
		}
		if len(anns) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No annotations found.")
			return nil
		}
		for k, v := range anns {
			fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", k, v)
		}
		return nil
	}

	annotations := map[string]string{}
	for _, pair := range args[1:] {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid annotation format %q: expected key=value", pair)
		}
		annotations[parts[0]] = parts[1]
	}

	if err := vault.AnnotateSecrets(cmd.Context(), client, path, annotations, annotateDryRun); err != nil {
		return fmt.Errorf("annotate failed: %w", err)
	}

	if annotateDryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "[dry-run] annotations not written")
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Annotated %s with %d key(s)\n", path, len(annotations))
	}
	return nil
}
