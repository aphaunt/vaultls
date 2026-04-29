package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Snapshot represents a point-in-time capture of all secrets at a path.
type Snapshot struct {
	Path      string            `json:"path"`
	Timestamp time.Time         `json:"timestamp"`
	Secrets   map[string]string `json:"secrets"`
}

// TakeSnapshot reads all secrets at the given path and returns a Snapshot.
func TakeSnapshot(ctx context.Context, client *Client, path string) (*Snapshot, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	keys, err := client.List(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("listing keys at %q: %w", path, err)
	}

	secrets := make(map[string]string, len(keys))
	for _, key := range keys {
		full := path + "/" + key
		data, err := client.Read(ctx, full)
		if err != nil {
			return nil, fmt.Errorf("reading secret %q: %w", full, err)
		}
		for k, v := range data {
			if s, ok := v.(string); ok {
				secrets[key+"."+k] = s
			}
		}
	}

	return &Snapshot{
		Path:      path,
		Timestamp: time.Now().UTC(),
		Secrets:   secrets,
	}, nil
}

// SnapshotToJSON serializes a Snapshot to indented JSON bytes.
func SnapshotToJSON(s *Snapshot) ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// DiffSnapshots compares two snapshots and returns a DiffResult.
func DiffSnapshots(a, b *Snapshot) map[string][2]string {
	result := make(map[string][2]string)

	for k, va := range a.Secrets {
		if vb, ok := b.Secrets[k]; !ok {
			result[k] = [2]string{va, ""}
		} else if va != vb {
			result[k] = [2]string{va, vb}
		}
	}
	for k, vb := range b.Secrets {
		if _, ok := a.Secrets[k]; !ok {
			result[k] = [2]string{"", vb}
		}
	}
	return result
}
