package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"vaultls/internal/vault"
)

var pinCmd = &cobra.Command{
	Use:   "pin",
	Short: "Pin, unpin, or check the pinned version of a secret",
}

var pinSetCmd = &cobra.Command{
	Use:   "set <path> <version>",
	Short: "Pin a secret to a specific version",
	Args:  cobra.ExactArgs(2),
	RunE:  runPinSet,
}

var pinGetCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Show the pinned version of a secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runPinGet,
}

var pinRemoveCmd = &cobra.Command{
	Use:   "remove <path>",
	Short: "Remove the pinned version from a secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runPinRemove,
}

func init() {
	pinCmd.AddCommand(pinSetCmd)
	pinCmd.AddCommand(pinGetCmd)
	pinCmd.AddCommand(pinRemoveCmd)
	rootCmd.AddCommand(pinCmd)
}

func runPinSet(cmd *cobra.Command, args []string) error {
	path := args[0]
	version, err := strconv.Atoi(args[1])
	if err != nil || version < 1 {
		return fmt.Errorf("version must be a positive integer")
	}

	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	if err := vault.PinSecret(cmd.Context(), client, path, version); err != nil {
		return fmt.Errorf("pin failed: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Pinned %s to version %d\n", path, version)
	return nil
}

func runPinGet(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	v, err := vault.GetPin(cmd.Context(), client, args[0])
	if err != nil {
		return fmt.Errorf("get pin failed: %w", err)
	}
	if v == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "%s is not pinned\n", args[0])
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "%s is pinned to version %d\n", args[0], v)
	}
	return nil
}

func runPinRemove(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}

	if err := vault.UnpinSecret(cmd.Context(), client, args[0]); err != nil {
		return fmt.Errorf("unpin failed: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Unpinned %s\n", args[0])
	return nil
}
