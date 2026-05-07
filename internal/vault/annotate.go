package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// AnnotateSecrets adds or updates metadata annotations on a KV v2 secret path.
// Annotations are stored as custom_metadata entries prefixed with "annotation/".
func AnnotateSecrets(ctx context.Context, client *vaultapi.Client, path string, annotations map[string]string, dryRun bool) error {
	if path == "" {
		return fmt.Errorf("path must not be empty")
	}
	if len(annotations) == 0 {
		return fmt.Errorf("at least one annotation must be provided")
	}

	mount, subPath := extractAnnotateMount(path)
	metaPath := fmt.Sprintf("%s/metadata/%s", mount, subPath)

	// Read existing metadata
	secret, err := client.Logical().ReadWithContext(ctx, metaPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata at %q: %w", metaPath, err)
	}

	existing := map[string]interface{}{}
	if secret != nil && secret.Data != nil {
		if cm, ok := secret.Data["custom_metadata"].(map[string]interface{}); ok {
			for k, v := range cm {
				existing[k] = v
			}
		}
	}

	for k, v := range annotations {
		existing["annotation/"+k] = v
	}

	if dryRun {
		return nil
	}

	_, err = client.Logical().WriteWithContext(ctx, metaPath, map[string]interface{}{
		"custom_metadata": existing,
	})
	if err != nil {
		return fmt.Errorf("failed to write annotations to %q: %w", metaPath, err)
	}
	return nil
}

// GetAnnotations retrieves annotation entries from a KV v2 secret's custom_metadata.
func GetAnnotations(ctx context.Context, client *vaultapi.Client, path string) (map[string]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	mount, subPath := extractAnnotateMount(path)
	metaPath := fmt.Sprintf("%s/metadata/%s", mount, subPath)

	secret, err := client.Logical().ReadWithContext(ctx, metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata at %q: %w", metaPath, err)
	}

	result := map[string]string{}
	if secret == nil || secret.Data == nil {
		return result, nil
	}

	cm, ok := secret.Data["custom_metadata"].(map[string]interface{})
	if !ok {
		return result, nil
	}

	for k, v := range cm {
		if strings.HasPrefix(k, "annotation/") {
			key := strings.TrimPrefix(k, "annotation/")
			result[key] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}

func extractAnnotateMount(path string) (string, string) {
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
