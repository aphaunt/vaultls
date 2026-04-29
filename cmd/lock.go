package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/vaultls/internal/vault"
)

var lockReason string

var lockCmd = &cobra.Command{
	Use:   "lock <path> <owner>",
	Short: "Lock a secret path to prevent modifications",
	Args:  cobra.ExactArgs(2),
	RunE:  runLock,
}

var unlockCmd = &cobra.Command{
	Use:   "unlock <path> <owner>",
	Short: "Unlock a previously locked secret path",
	Args:  cobra.ExactArgs(2),
	RunE:  runUnlock,
}

var lockStatusCmd = &cobra.Command{
	Use:   "lock-status <path>",
	Short: "Show lock status for a secret path",
	Args:  cobra.ExactArgs(1),
	RunE:  runLockStatus,
}

func init() {
	lockCmd.Flags().StringVarP(&lockReason, "reason", "r", "", "Reason for locking the secret")
	rootCmd.AddCommand(lockCmd)
	rootCmd.AddCommand(unlockCmd)
	rootCmd.AddCommand(lockStatusCmd)
}

func runLock(cmd *cobra.Command, args []string) error {
	path, owner := args[0], args[1]
	client, err := vault.NewClient(vaultAddr, vaultToken)
	if err != nil {
		return fmt.Errorf("client error: %w", err)
	}
	if err := vault.LockSecret(client, path, owner, lockReason); err != nil {
		return fmt.Errorf("lock failed: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Locked %s (owner: %s)\n", path, owner)
	return nil
}

func runUnlock(cmd *cobra.Command, args []string) error {
	path, owner := args[0], args[1]
	client, err := vault.NewClient(vaultAddr, vaultToken)
	if err != nil {
		return fmt.Errorf("client error: %w", err)
	}
	if err := vault.UnlockSecret(client, path, owner); err != nil {
		return fmt.Errorf("unlock failed: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Unlocked %s\n", path)
	return nil
}

func runLockStatus(cmd *cobra.Command, args []string) error {
	path := args[0]
	client, err := vault.NewClient(vaultAddr, vaultToken)
	if err != nil {
		return fmt.Errorf("client error: %w", err)
	}
	lock, err := vault.GetLock(client, path)
	if err != nil {
		return fmt.Errorf("status check failed: %w", err)
	}
	if lock == nil {
		fmt.Fprintf(cmd.OutOrStdout(), "%s is not locked\n", path)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Locked by: %s\nReason: %s\nSince: %s\n",
		lock.LockedBy, lock.Reason, lock.LockedAt.Format("2006-01-02 15:04:05"))
	return nil
}
