package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func executeBookmarkCmd(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"bookmark"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestBookmarkCmd_AddMissingArgs(t *testing.T) {
	_, err := executeBookmarkCmd("add")
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}

func TestBookmarkCmd_AddTooFewArgs(t *testing.T) {
	_, err := executeBookmarkCmd("add", "mykey")
	if err == nil {
		t.Fatal("expected error for too few args")
	}
}

func TestBookmarkCmd_RemoveMissingArgs(t *testing.T) {
	_, err := executeBookmarkCmd("remove")
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}

func TestBookmarkCmd_UsageText(t *testing.T) {
	out, _ := executeBookmarkCmd("--help")
	if !strings.Contains(out, "bookmark") {
		t.Errorf("expected usage to mention 'bookmark', got: %s", out)
	}
}

func TestBookmarkCmd_NoteFlag(t *testing.T) {
	out, _ := executeBookmarkCmd("add", "--help")
	if !strings.Contains(out, "--note") {
		t.Errorf("expected --note flag in help output, got: %s", out)
	}
}
