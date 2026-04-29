package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockSnapshotServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/app":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"db", "api"}},
			})
		case "/v1/secret/data/app/db":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"password": "s3cr3t"}},
			})
		case "/v1/secret/data/app/api":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"key": "abc123"}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestTakeSnapshot_EmptyPath(t *testing.T) {
	client := &Client{}
	_, err := TakeSnapshot(context.Background(), client, "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSnapshotToJSON_RoundTrip(t *testing.T) {
	snap := &Snapshot{
		Path:    "secret/app",
		Secrets: map[string]string{"db.password": "s3cr3t"},
	}
	b, err := SnapshotToJSON(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out Snapshot
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if out.Path != snap.Path {
		t.Errorf("expected path %q, got %q", snap.Path, out.Path)
	}
	if out.Secrets["db.password"] != "s3cr3t" {
		t.Errorf("expected secret value s3cr3t, got %q", out.Secrets["db.password"])
	}
}

func TestDiffSnapshots_DetectsChanges(t *testing.T) {
	a := &Snapshot{Secrets: map[string]string{"key1": "val1", "key2": "old"}}
	b := &Snapshot{Secrets: map[string]string{"key2": "new", "key3": "added"}}

	diff := DiffSnapshots(a, b)

	if diff["key1"] != ([2]string{"val1", ""}) {
		t.Errorf("expected key1 removed, got %v", diff["key1"])
	}
	if diff["key2"] != ([2]string{"old", "new"}) {
		t.Errorf("expected key2 changed, got %v", diff["key2"])
	}
	if diff["key3"] != ([2]string{"", "added"}) {
		t.Errorf("expected key3 added, got %v", diff["key3"])
	}
}

func TestDiffSnapshots_NoChanges(t *testing.T) {
	a := &Snapshot{Secrets: map[string]string{"k": "v"}}
	b := &Snapshot{Secrets: map[string]string{"k": "v"}}
	diff := DiffSnapshots(a, b)
	if len(diff) != 0 {
		t.Errorf("expected no diff, got %v", diff)
	}
}
