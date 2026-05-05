package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultls/internal/vault"
)

var bookmarkNoteFlag string

func init() {
	bookmarkAddCmd := &cobra.Command{
		Use:   "add <name> <path>",
		Short: "Save a bookmark to a Vault path",
		Args:  cobra.ExactArgs(2),
		RunE:  runBookmarkAdd,
	}
	bookmarkAddCmd.Flags().StringVar(&bookmarkNoteFlag, "note", "", "Optional note for the bookmark")

	bookmarkRemoveCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a saved bookmark",
		Args:  cobra.ExactArgs(1),
		RunE:  runBookmarkRemove,
	}

	bookmarkListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved bookmarks",
		Args:  cobra.NoArgs,
		RunE:  runBookmarkList,
	}

	bookmarkCmd := &cobra.Command{
		Use:   "bookmark",
		Short: "Manage bookmarks for frequently used Vault paths",
	}
	bookmarkCmd.AddCommand(bookmarkAddCmd, bookmarkRemoveCmd, bookmarkListCmd)
	rootCmd.AddCommand(bookmarkCmd)
}

func runBookmarkAdd(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient()
	if err != nil {
		return err
	}
	if err := vault.AddBookmark(cmd.Context(), client, args[0], args[1], bookmarkNoteFlag); err != nil {
		return fmt.Errorf("add bookmark: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Bookmark %q -> %s saved\n", args[0], args[1])
	return nil
}

func runBookmarkRemove(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient()
	if err != nil {
		return err
	}
	if err := vault.RemoveBookmark(cmd.Context(), client, args[0]); err != nil {
		return fmt.Errorf("remove bookmark: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Bookmark %q removed\n", args[0])
	return nil
}

func runBookmarkList(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient()
	if err != nil {
		return err
	}
	bookmarks, err := vault.ListBookmarks(cmd.Context(), client)
	if err != nil {
		return fmt.Errorf("list bookmarks: %w", err)
	}
	if len(bookmarks) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No bookmarks saved.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPATH\tNOTE")
	for _, b := range bookmarks {
		fmt.Fprintf(w, "%s\t%s\t%s\n", b.Name, b.Path, b.Note)
	}
	return w.Flush()
}
