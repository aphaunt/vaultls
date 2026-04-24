package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"gopkg.in/yaml.v3"
)

// ExportFormat represents the output format for secret export.
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatYAML ExportFormat = "yaml"
	FormatEnv  ExportFormat = "env"
)

// ExportSecrets writes the secrets map to the writer in the specified format.
func ExportSecrets(w io.Writer, secrets map[string]string, format ExportFormat) error {
	switch format {
	case FormatJSON:
		return exportJSON(w, secrets)
	case FormatYAML:
		return exportYAML(w, secrets)
	case FormatEnv:
		return exportEnv(w, secrets)
	default:
		return fmt.Errorf("unsupported format: %q", format)
	}
}

func exportJSON(w io.Writer, secrets map[string]string) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(secrets)
}

func exportYAML(w io.Writer, secrets map[string]string) error {
	return yaml.NewEncoder(w).Encode(secrets)
}

func exportEnv(w io.Writer, secrets map[string]string) error {
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "%s=%q\n", k, secrets[k]); err != nil {
			return err
		}
	}
	return nil
}
