package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func executeLockCmd(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestLockCmd_MissingArgs(t *testing.T) {
	_, err := executeLockCmd("lock")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg")
}

func TestLockCmd_TooFewArgs(t *testing.T) {
	_, err := executeLockCmd("lock", "secret/data/myapp")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg")
}

func TestLockCmd_UsageText(t *testing.T) {
	out, _ := executeLockCmd("lock", "--help")
	assert.Contains(t, out, "lock <path> <owner>")
	assert.Contains(t, out, "Lock a secret path")
}

func TestLockCmd_ReasonFlag(t *testing.T) {
	out, _ := executeLockCmd("lock", "--help")
	assert.Contains(t, out, "--reason")
	assert.Contains(t, out, "Reason for locking")
}

func TestUnlockCmd_MissingArgs(t *testing.T) {
	_, err := executeLockCmd("unlock")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg")
}

func TestUnlockCmd_UsageText(t *testing.T) {
	out, _ := executeLockCmd("unlock", "--help")
	assert.Contains(t, out, "unlock <path> <owner>")
	assert.Contains(t, out, "Unlock a previously locked")
}

func TestLockStatusCmd_MissingArgs(t *testing.T) {
	_, err := executeLockCmd("lock-status")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg")
}

func TestLockStatusCmd_UsageText(t *testing.T) {
	out, _ := executeLockCmd("lock-status", "--help")
	assert.Contains(t, out, "lock-status <path>")
	assert.Contains(t, out, "Show lock status")
}
