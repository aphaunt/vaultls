package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockDiffPathsServer(t *testing.T, responses map[string][]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		keys, ok := responses[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		iface := make([]interface{}, len(keys))
		for i, k := range keys {
			iface[i] = k
		}
		body := map[string]interface{}{"data": map[string]interface{}{"keys": iface}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
}

func newDiffPathsClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestDiffPaths_EmptyPathA(t *testing.T) {
	server := newMockDiffPathsServer(t, nil)
	defer server.Close()
	client := newDiffPathsClient(t, server.URL)
	_, err := DiffPaths(context.Background(), client, "", "secret/b")
	if err == nil {
		t.Fatal("expected error for empty pathA")
	}
}

func TestDiffPaths_EmptyPathB(t *testing.T) {
	server := newMockDiffPathsServer(t, nil)
	defer server.Close()
	client := newDiffPathsClient(t, server.URL)
	_, err := DiffPaths(context.Background(), client, "secret/a", "")
	if err == nil {
		t.Fatal("expected error for empty pathB")
	}
}

func TestDiffPaths_DisjointKeys(t *testing.T) {
	server := newMockDiffPathsServer(t, map[string][]string{
		"/v1/secret/a": {"foo", "bar"},
		"/v1/secret/b": {"baz", "qux"},
	})
	defer server.Close()
	client := newDiffPathsClient(t, server.URL)
	res, err := DiffPaths(context.Background(), client, "secret/a", "secret/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.InBoth) != 0 {
		t.Errorf("expected 0 shared keys, got %d", len(res.InBoth))
	}
	if len(res.OnlyInA) != 2 {
		t.Errorf("expected 2 keys only in A, got %d", len(res.OnlyInA))
	}
	if len(res.OnlyInB) != 2 {
		t.Errorf("expected 2 keys only in B, got %d", len(res.OnlyInB))
	}
}

func TestDiffPaths_OverlappingKeys(t *testing.T) {
	server := newMockDiffPathsServer(t, map[string][]string{
		"/v1/secret/a": {"foo", "bar", "shared"},
		"/v1/secret/b": {"baz", "shared"},
	})
	defer server.Close()
	client := newDiffPathsClient(t, server.URL)
	res, err := DiffPaths(context.Background(), client, "secret/a", "secret/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.InBoth) != 1 || res.InBoth[0] != "shared" {
		t.Errorf("expected [shared] in both, got %v", res.InBoth)
	}
	if len(res.OnlyInA) != 2 {
		t.Errorf("expected 2 only in A, got %d", len(res.OnlyInA))
	}
	if len(res.OnlyInB) != 1 {
		t.Errorf("expected 1 only in B, got %d", len(res.OnlyInB))
	}
}

func TestDiffPaths_IdenticalKeys(t *testing.T) {
	server := newMockDiffPathsServer(t, map[string][]string{
		"/v1/secret/a": {"x", "y"},
		"/v1/secret/b": {"x", "y"},
	})
	defer server.Close()
	client := newDiffPathsClient(t, server.URL)
	res, err := DiffPaths(context.Background(), client, "secret/a", "secret/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.InBoth) != 2 {
		t.Errorf("expected 2 shared keys, got %d", len(res.InBoth))
	}
	if len(res.OnlyInA) != 0 || len(res.OnlyInB) != 0 {
		t.Errorf("expected no unique keys, got A=%v B=%v", res.OnlyInA, res.OnlyInB)
	}
}
