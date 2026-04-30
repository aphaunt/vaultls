package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

// TemplateResult holds the rendered output for a single secret path.
type TemplateResult struct {
	Path     string
	Rendered string
}

// RenderTemplate reads secrets from the given path and substitutes
// {{KEY}} placeholders in the provided template string.
func RenderTemplate(ctx context.Context, client *api.Client, path, tmpl string) (*TemplateResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	if tmpl == "" {
		return nil, fmt.Errorf("template must not be empty")
	}

	secret, err := client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading path %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path %q", path)
	}

	data, ok := secret.Data["data"]
	if !ok {
		data = secret.Data
	}

	kv, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at path %q", path)
	}

	result := tmpl
	for k, v := range kv {
		placeholder := "{{" + k + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", v))
	}

	return &TemplateResult{Path: path, Rendered: result}, nil
}
