package vault

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// RenderDiff writes a human-readable diff to w.
func RenderDiff(w io.Writer, label string, diffs []DiffResult) {
	if len(diffs) == 0 {
		fmt.Fprintf(w, "No differences found for %s\n", label)
		return
	}

	fmt.Fprintf(w, "Diff for %s:\n", label)
	fmt.Fprintln(w, strings.Repeat("-", 40))

	sorted := make([]DiffResult, len(diffs))
	copy(sorted, diffs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})

	for _, d := range sorted {
		switch {
		case d.ValueA == "" && d.ValueB != "":
			fmt.Fprintf(w, "%s  %s: (missing) -> %q\n", colorize("+", "green"), d.Key, d.ValueB)
		case d.ValueA != "" && d.ValueB == "":
			fmt.Fprintf(w, "%s  %s: %q -> (missing)\n", colorize("-", "red"), d.Key, d.ValueA)
		default:
			fmt.Fprintf(w, "%s  %s: %q -> %q\n", colorize("~", "yellow"), d.Key, d.ValueA, d.ValueB)
		}
	}
}

// colorize wraps text in ANSI color codes.
func colorize(text, color string) string {
	codes := map[string]string{
		"red":    "\033[31m",
		"green":  "\033[32m",
		"yellow": "\033[33m",
	}
	reset := "\033[0m"
	if code, ok := codes[color]; ok {
		return code + text + reset
	}
	return text
}
