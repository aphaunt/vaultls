package vault

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"
)

// LockEntry represents a lock placed on a secret path.
type LockEntry struct {
	Path      string    `json:"path"`
	Owner     string    `json:"owner"`
	Reason    string    `json:"reason"`
	LockedAt  time.Time `json:"locked_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// lockMetaKey returns the metadata key used to store lock information.
func lockMetaKey(secretPath string) string {
	clean := strings.TrimSuffix(secretPath, "/")
	dir := path.Dir(clean)
	base := path.Base(clean)
	return path.Join(dir, ".locks", base)
}

// LockSecret places a lock on a secret path by writing lock metadata.
// If the path is already locked and overwrite is false, an error is returned.
func LockSecret(ctx context.Context, client *Client, secretPath, owner, reason string, ttl time.Duration, overwrite bool) (*LockEntry, error) {
	if strings.TrimSpace(secretPath) == "" {
		return nil, fmt.Errorf("secret path must not be empty")
	}
	if strings.TrimSpace(owner) == "" {
		return nil, fmt.Errorf("lock owner must not be empty")
	}

	metaPath := lockMetaKey(secretPath)

	// Check for an existing lock.
	existing, err := client.GetSecret(ctx, metaPath)
	if err == nil && existing != nil {
		if !overwrite {
			return nil, fmt.Errorf("path %q is already locked by %q; use --overwrite to replace", secretPath, existing["owner"])
		}
	}

	now := time.Now().UTC()
	entry := &LockEntry{
		Path:     secretPath,
		Owner:    owner,
		Reason:   reason,
		LockedAt: now,
	}
	if ttl > 0 {
		entry.ExpiresAt = now.Add(ttl)
	}

	data := map[string]interface{}{
		"path":       entry.Path,
		"owner":      entry.Owner,
		"reason":     entry.Reason,
		"locked_at":  entry.LockedAt.Format(time.RFC3339),
		"expires_at": entry.ExpiresAt.Format(time.RFC3339),
	}

	if err := client.WriteSecret(ctx, metaPath, data); err != nil {
		return nil, fmt.Errorf("failed to write lock metadata: %w", err)
	}

	return entry, nil
}

// UnlockSecret removes a lock from a secret path.
// Returns an error if the path is not currently locked.
func UnlockSecret(ctx context.Context, client *Client, secretPath string) error {
	if strings.TrimSpace(secretPath) == "" {
		return fmt.Errorf("secret path must not be empty")
	}

	metaPath := lockMetaKey(secretPath)

	existing, err := client.GetSecret(ctx, metaPath)
	if err != nil || existing == nil {
		return fmt.Errorf("path %q is not currently locked", secretPath)
	}

	if err := client.DeleteSecret(ctx, metaPath); err != nil {
		return fmt.Errorf("failed to remove lock metadata: %w", err)
	}

	return nil
}

// GetLock retrieves the current lock entry for a secret path.
// Returns nil if the path is not locked.
func GetLock(ctx context.Context, client *Client, secretPath string) (*LockEntry, error) {
	if strings.TrimSpace(secretPath) == "" {
		return nil, fmt.Errorf("secret path must not be empty")
	}

	metaPath := lockMetaKey(secretPath)

	data, err := client.GetSecret(ctx, metaPath)
	if err != nil || data == nil {
		return nil, nil
	}

	entry := &LockEntry{}
	if v, ok := data["path"].(string); ok {
		entry.Path = v
	}
	if v, ok := data["owner"].(string); ok {
		entry.Owner = v
	}
	if v, ok := data["reason"].(string); ok {
		entry.Reason = v
	}
	if v, ok := data["locked_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			entry.LockedAt = t
		}
	}
	if v, ok := data["expires_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil && !t.IsZero() {
			entry.ExpiresAt = t
		}
	}

	// Treat expired locks as absent.
	if !entry.ExpiresAt.IsZero() && time.Now().UTC().After(entry.ExpiresAt) {
		return nil, nil
	}

	return entry, nil
}
