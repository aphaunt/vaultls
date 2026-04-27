package vault

import (
	"context"
	"fmt"
	"time"
)

// WatchResult holds the result of a single poll during a watch operation.
type WatchResult struct {
	Path      string
	Changes   []SecretDiff
	Timestamp time.Time
	Error     error
}

// WatchOptions configures the behavior of WatchSecrets.
type WatchOptions struct {
	// Interval between polls. Defaults to 10 seconds if zero.
	Interval time.Duration
	// MaxChanges stops the watcher after this many change events (0 = unlimited).
	MaxChanges int
}

// WatchSecrets polls a Vault KV path at a regular interval and emits a
// WatchResult on the returned channel whenever the secret data changes.
// The caller is responsible for cancelling the context to stop watching.
func WatchSecrets(ctx context.Context, client *Client, path string, opts WatchOptions) (<-chan WatchResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	interval := opts.Interval
	if interval <= 0 {
		interval = 10 * time.Second
	}

	resultCh := make(chan WatchResult, 1)

	go func() {
		defer close(resultCh)

		// Capture initial state.
		prev, err := client.GetSecrets(ctx, path)
		if err != nil {
			resultCh <- WatchResult{Path: path, Error: err, Timestamp: time.Now()}
			return
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		changeCount := 0

		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				curr, err := client.GetSecrets(ctx, path)
				if err != nil {
					resultCh <- WatchResult{Path: path, Error: err, Timestamp: t}
					continue
				}

				diffs := DiffSecrets(prev, curr)
				if len(diffs) == 0 {
					continue
				}

				resultCh <- WatchResult{
					Path:      path,
					Changes:   diffs,
					Timestamp: t,
				}

				prev = curr
				changeCount++

				if opts.MaxChanges > 0 && changeCount >= opts.MaxChanges {
					return
				}
			}
		}
	}()

	return resultCh, nil
}
