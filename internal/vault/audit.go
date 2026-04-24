package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AuditEntry represents a single audit log entry from Vault.
type AuditEntry struct {
	Time      time.Time         `json:"time"`
	Type      string            `json:"type"`
	Path      string            `json:"path"`
	Operation string            `json:"operation"`
	Metadata  map[string]string `json:"metadata"`
}

// AuditLog holds a list of audit entries.
type AuditLog struct {
	Entries []AuditEntry
}

// GetAuditLog fetches recent audit log entries for a given path from Vault.
func GetAuditLog(ctx context.Context, client *Client, path string) (*AuditLog, error) {
	url := fmt.Sprintf("%s/v1/sys/audit-hash/%s", client.Address, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building audit request: %w", err)
	}
	req.Header.Set("X-Vault-Token", client.Token)

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing audit request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from audit endpoint", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Entries []AuditEntry `json:"entries"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding audit response: %w", err)
	}

	return &AuditLog{Entries: result.Data.Entries}, nil
}

// FilterByOperation returns entries matching the given operation type (e.g. "read", "write").
func (a *AuditLog) FilterByOperation(op string) []AuditEntry {
	var filtered []AuditEntry
	for _, e := range a.Entries {
		if e.Operation == op {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
