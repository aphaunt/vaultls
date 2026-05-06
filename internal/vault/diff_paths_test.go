package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMockDiffPathsServer(t *testing.T, responses map[string][]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys, ok := responses[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		ifaces := make([]interface{}, len(keys))
		for i, k := range keys {
			ifaces[i] = k
		}
		body := map[string]interface{}{"data": map[string]interface{}{"keys": ifaces}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
}

func newDiffPathsClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestDiffPaths_EmptyPathA(t *testing.T) {
	srv := newMockDiffPathsServer(t, map[string][]string{})
	defer srv.Close()
	client := newDiffPathsClient(t, srv)

	_, err := DiffPaths(context.Background(), client, "", "secret/b")
	if err == nil {
		t.Fatal("expected error for empty pathA")
	}
}

func TestDiffPaths_EmptyPathB(t *testing.T) {
	srv := newMockDiffPathsServer(t, map[string][]string{})
	defer srv.Close()
	client := newDiffPathsClient(t, srv)

	_, err := DiffPaths(context.Background(), client, "secret/a", "")
	if err == nil {
		t.Fatal("expected error for empty pathB")
	}
}

func TestDiffPaths_DisjointKeys(t *testing.T) {
	srv := newMockDiffPathsServer(t, map[string][]string{
		"/v1/secret/a": {"foo", "bar"},
		"/v1/secret/b": {"baz", "qux"},
	})
	defer srv.Close()
	client := newDiffPathsClient(t, srv)

	diff, err := DiffPaths(context.Background(), client, "secret/a", "secret/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diff.InBoth) != 0 {
		t.Errorf("expected no shared keys, got %v", diff.InBoth)
	}
	if len(diff.OnlyInA) != 2 {
		t.Errorf("expected 2 keys only in A, got %v", diff.OnlyInA)
	}
	if len(diff.OnlyInB) != 2 {
		t.Errorf("expected 2 keys only in B, got %v", diff.OnlyInB)
	}
}

func TestDiffPaths_OverlappingKeys(t *testing.T) {
	srv := newMockDiffPathsServer(t, map[string][]string{
		"/v1/secret/a": {"shared", "only-a"},
		"/v1/secret/b": {"shared", "only-b"},
	})
	defer srv.Close()
	client := newDiffPathsClient(t, srv)

	diff, err := DiffPaths(context.Background(), client, "secret/a", "secret/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diff.InBoth) != 1 || diff.InBoth[0] != "shared" {
		t.Errorf("expected [shared] in both, got %v", diff.InBoth)
	}
	if len(diff.OnlyInA) != 1 || diff.OnlyInA[0] != "only-a" {
		t.Errorf("expected [only-a], got %v", diff.OnlyInA)
	}
	if len(diff.OnlyInB) != 1 || diff.OnlyInB[0] != "only-b" {
		t.Errorf("expected [only-b], got %v", diff.OnlyInB)
	}
}

func TestDiffPaths_IdenticalKeys(t *testing.T) {
	srv := newMockDiffPathsServer(t, map[string][]string{
		"/v1/secret/a": {"x", "y"},
		"/v1/secret/b": {"x", "y"},
	})
	defer srv.Close()
	client := newDiffPathsClient(t, srv)

	diff, err := DiffPaths(context.Background(), client, "secret/a", "secret/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diff.OnlyInA) != 0 || len(diff.OnlyInB) != 0 {
		t.Errorf("expected no unique keys, got onlyA=%v onlyB=%v", diff.OnlyInA, diff.OnlyInB)
	}
	if len(diff.InBoth) != 2 {
		t.Errorf("expected 2 shared keys, got %v", diff.InBoth)
	}
}
