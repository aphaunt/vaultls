package vault

import (
	"context"
	"fmt"
	"path"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// ArchiveResult holds the outcome of an archive operation.
type ArchiveResult struct {
	Path    string
	Archived bool
	Skipped  bool
	Reason   string
}

// ArchiveSecrets moves secrets from src to an archive destination by
// writing them under archiveRoot and optionally deleting the originals.
func ArchiveSecrets(ctx context.Context, client *vaultapi.Client, src, archiveRoot string, deleteOriginal bool, dryRun bool) ([]ArchiveResult, error) {
	if strings.TrimSpace(src) == "" {
		return nil, fmt.Errorf("source path must not be empty")
	}
	if strings.TrimSpace(archiveRoot) == "" {
		return nil, fmt.Errorf("archive root must not be empty")
	}
	if src == archiveRoot {
		return nil, fmt.Errorf("source and archive root must differ")
	}

	listPath := "secret/metadata/" + strings.TrimPrefix(src, "/")
	secret, err := client.Logical().ListWithContext(ctx, listPath)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", src, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no secrets found at %s", src)
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected keys format at %s", src)
	}

	var results []ArchiveResult
	for _, k := range keys {
		key, _ := k.(string)
		if strings.HasSuffix(key, "/") {
			continue
		}

		srcFull := path.Join("secret/data", src, key)
		dstFull := path.Join("secret/data", archiveRoot, key)

		if dryRun {
			results = append(results, ArchiveResult{Path: srcFull, Archived: false, Skipped: true, Reason: "dry-run"})
			continue
		}

		readSecret, err := client.Logical().ReadWithContext(ctx, srcFull)
		if err != nil || readSecret == nil {
			results = append(results, ArchiveResult{Path: srcFull, Skipped: true, Reason: "read error"})
			continue
		}

		data := map[string]interface{}{"data": readSecret.Data["data"]}
		_, err = client.Logical().WriteWithContext(ctx, dstFull, data)
		if err != nil {
			results = append(results, ArchiveResult{Path: srcFull, Skipped: true, Reason: "write error: " + err.Error()})
			continue
		}

		if deleteOriginal {
			delPath := path.Join("secret/metadata", src, key)
			_, _ = client.Logical().DeleteWithContext(ctx, delPath)
		}

		results = append(results, ArchiveResult{Path: srcFull, Archived: true})
	}

	return results, nil
}
