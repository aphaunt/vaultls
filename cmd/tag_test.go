package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeTagCmd(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"tag"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestTagCmd_MissingArgs(t *testing.T) {
	_, err := executeTagCmd()
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestTagCmd_MissingTagPairs(t *testing.T) {
	_, err := executeTagCmd("secret/myapp/config")
	if err == nil {
		t.Fatal("expected error when no key=value pairs provided")
	}
}

func TestTagCmd_InvalidTagFormat(t *testing.T) {
	cmd := &cobra.Command{}
	err := runTag(cmd, []string{"secret/myapp/config", "badformat"})
	if err == nil || !strings.Contains(err.Error(), "invalid tag format") {
		t.Fatalf("expected invalid tag format error, got: %v", err)
	}
}

func TestTagCmd_UsageText(t *testing.T) {
	out, _ := executeTagCmd("--help")
	if !strings.Contains(out, "tag") {
		t.Errorf("expected 'tag' in usage output, got: %s", out)
	}
	if !strings.Contains(out, "key=value") {
		t.Errorf("expected 'key=value' in usage output, got: %s", out)
	}
}

func TestTagCmd_OverwriteFlag(t *testing.T) {
	cmd := tagCmd
	if cmd.Flags().Lookup("overwrite") == nil {
		t.Error("expected --overwrite flag to be defined")
	}
}

func TestTagCmd_ListFlag(t *testing.T) {
	cmd := tagCmd
	if cmd.Flags().Lookup("list") == nil {
		t.Error("expected --list flag to be defined")
	}
}
