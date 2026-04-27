package vault

import (
	"context"
	"fmt"
	"io"
	"time"
)

// WatchOptions configures the watch behavior.
type WatchOptions struct {
	Interval time.Duration
	MaxChanges int
	Out        io.Writer
}

// WatchSecretsWithOutput watches a Vault path and writes change events to the
// provided writer, returning after MaxChanges events or when ctx is cancelled.
func WatchSecretsWithOutput(ctx context.Context, client *Client, path string, opts WatchOptions) error {
	if path == "" {
		return fmt.Errorf("path must not be empty")
	}
	if opts.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	if opts.Out == nil {
		return fmt.Errorf("output writer must not be nil")
	}

	changes := 0
	events, err := WatchSecrets(ctx, client, path, opts.Interval)
	if err != nil {
		return fmt.Errorf("failed to start watch: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-events:
			if !ok {
				return nil
			}
			fmt.Fprintf(opts.Out, "[%s] %s: %s\n",
				ev.Timestamp.Format(time.RFC3339), ev.Key, ev.ChangeType)
			changes++
			if opts.MaxChanges > 0 && changes >= opts.MaxChanges {
				return nil
			}
		}
	}
}
