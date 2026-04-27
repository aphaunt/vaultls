package vault_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultls/internal/vault"
)

func newMockWatchServer(t *testing.T, responses []map[string]interface{}) *httptest.Server {
	t.Helper()
	call := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := call
		if idx >= len(responses) {
			idx = len(responses) - 1
		}
		call++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": responses[idx],
		})
	}))
}

func TestWatchSecretsWithOutput_EmptyPath(t *testing.T) {
	client, _ := vault.NewClient("http://localhost", "token")
	var buf bytes.Buffer
	err := vault.WatchSecretsWithOutput(context.Background(), client, "", vault.WatchOptions{
		Interval: time.Second,
		Out:      &buf,
	})
	if err == nil || !strings.Contains(err.Error(), "path") {
		t.Errorf("expected path error, got %v", err)
	}
}

func TestWatchSecretsWithOutput_InvalidInterval(t *testing.T) {
	client, _ := vault.NewClient("http://localhost", "token")
	var buf bytes.Buffer
	err := vault.WatchSecretsWithOutput(context.Background(), client, "secret/app", vault.WatchOptions{
		Interval: 0,
		Out:      &buf,
	})
	if err == nil || !strings.Contains(err.Error(), "interval") {
		t.Errorf("expected interval error, got %v", err)
	}
}

func TestWatchSecretsWithOutput_ContextCancelled(t *testing.T) {
	responses := []map[string]interface{}{
		{"KEY": "val1"},
	}
	svr := newMockWatchServer(t, responses)
	defer svr.Close()

	client, err := vault.NewClient(svr.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	err = vault.WatchSecretsWithOutput(ctx, client, "secret/app", vault.WatchOptions{
		Interval: 10 * time.Millisecond,
		Out:      &buf,
	})
	if err != nil {
		t.Errorf("unexpected error on cancelled context: %v", err)
	}
	_ = fmt.Sprintf("output: %s", buf.String())
}
