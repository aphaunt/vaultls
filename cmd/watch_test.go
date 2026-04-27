package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeWatchCmd(args []string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"watch"}, args...))
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestWatchCmd_MissingArgs(t *testing.T) {
	_, err := executeWatchCmd([]string{})
	if err == nil {
		t.Error("expected error for missing path argument")
	}
}

func TestWatchCmd_TooManyArgs(t *testing.T) {
	_, err := executeWatchCmd([]string{"secret/a", "secret/b"})
	if err == nil {
		t.Error("expected error for too many arguments")
	}
}

func TestWatchCmd_UsageText(t *testing.T) {
	cmd := &cobra.Command{}
	watchCmd.SetOut(cmd.OutOrStdout())
	use := watchCmd.Use
	if !strings.HasPrefix(use, "watch") {
		t.Errorf("expected Use to start with 'watch', got %q", use)
	}
}

func TestWatchCmd_IntervalFlag(t *testing.T) {
	f := watchCmd.Flags().Lookup("interval")
	if f == nil {
		t.Fatal("expected --interval flag to be registered")
	}
	if f.DefValue != "5s" {
		t.Errorf("expected default interval 5s, got %s", f.DefValue)
	}
}

func TestWatchCmd_MaxChangesFlag(t *testing.T) {
	f := watchCmd.Flags().Lookup("max-changes")
	if f == nil {
		t.Fatal("expected --max-changes flag to be registered")
	}
	if f.DefValue != "0" {
		t.Errorf("expected default max-changes 0, got %s", f.DefValue)
	}
}
