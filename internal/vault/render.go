package vault

import (
	"fmt"
	"io"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// RenderDiff writes a human-readable diff of a DiffResult to the given writer.
func RenderDiff(w io.Writer, result DiffResult, envA, envB string, useColor bool) {
	if len(result.OnlyInA) == 0 && len(result.OnlyInB) == 0 && len(result.Changed) == 0 {
		fmt.Fprintln(w, "No differences found.")
		return
	}

	for _, k := range result.OnlyInA {
		line := fmt.Sprintf("- [%s only] %s", envA, k)
		fmt.Fprintln(w, colorize(line, colorRed, useColor))
	}

	for _, k := range result.OnlyInB {
		line := fmt.Sprintf("+ [%s only] %s", envB, k)
		fmt.Fprintln(w, colorize(line, colorGreen, useColor))
	}

	for _, k := range result.InBoth {
		if vals, ok := result.Changed[k]; ok {
			sep := strings.Repeat("-", 40)
			fmt.Fprintln(w, colorize(sep, colorYellow, useColor))
			fmt.Fprintln(w, colorize(fmt.Sprintf("~ %s", k), colorYellow, useColor))
			fmt.Fprintln(w, colorize(fmt.Sprintf("  %s: %s", envA, vals[0]), colorRed, useColor))
			fmt.Fprintln(w, colorize(fmt.Sprintf("  %s: %s", envB, vals[1]), colorGreen, useColor))
		}
	}
}

func colorize(s, color string, useColor bool) string {
	if !useColor {
		return s
	}
	return color + s + colorReset
}
