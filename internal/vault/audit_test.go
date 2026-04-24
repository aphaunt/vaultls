package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newMockAuditServer(t *testing.T, entries []AuditEntry) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"entries": entries,
			},
		})
	}))
}

func TestGetAuditLog_Success(t *testing.T) {
	entries := []AuditEntry{
		{Time: time.Now(), Type: "request", Path: "secret/foo", Operation: "read"},
		{Time: time.Now(), Type: "request", Path: "secret/foo", Operation: "write"},
	}
	server := newMockAuditServer(t, entries)
	defer server.Close()

	client := &Client{Address: server.URL, Token: "test-token", HTTP: server.Client()}
	log, err := GetAuditLog(context.Background(), client, "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(log.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(log.Entries))
	}
}

func TestGetAuditLog_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{Address: server.URL, Token: "test-token", HTTP: server.Client()}
	_, err := GetAuditLog(context.Background(), client, "secret/foo")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFilterByOperation(t *testing.T) {
	log := &AuditLog{
		Entries: []AuditEntry{
			{Operation: "read"},
			{Operation: "write"},
			{Operation: "read"},
		},
	}
	reads := log.FilterByOperation("read")
	if len(reads) != 2 {
		t.Errorf("expected 2 read entries, got %d", len(reads))
	}
	writes := log.FilterByOperation("write")
	if len(writes) != 1 {
		t.Errorf("expected 1 write entry, got %d", len(writes))
	}
}

func TestFilterByOperation_NoMatch(t *testing.T) {
	log := &AuditLog{
		Entries: []AuditEntry{
			{Operation: "read"},
		},
	}
	result := log.FilterByOperation("delete")
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}
