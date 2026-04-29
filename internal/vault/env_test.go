package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockEnvServer(t *testing.T, dataA, dataB map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if strings.Contains(r.URL.Path, "pathA") {
			payload = map[string]interface{}{"data": dataA}
		} else {
			payload = map[string]interface{}{"data": dataB}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": payload})
	}))
}

func newEnvClient(t *testing.T, server *httptest.Server) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = server.URL
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestCompareEnvs_EmptyPaths(t *testing.T) {
	server := newMockEnvServer(t, nil, nil)
	defer server.Close()
	client := newEnvClient(t, server)

	_, err := CompareEnvs(context.Background(), client, "", "secret/pathB")
	if err == nil {
		t.Fatal("expected error for empty pathA")
	}
}

func TestCompareEnvs_Identical(t *testing.T) {
	data := map[string]interface{}{"KEY": "value"}
	server := newMockEnvServer(t, data, data)
	defer server.Close()
	client := newEnvClient(t, server)

	diff, err := CompareEnvs(context.Background(), client, "secret/pathA", "secret/pathB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diff.Changed) != 0 || len(diff.OnlyInA) != 0 || len(diff.OnlyInB) != 0 {
		t.Errorf("expected no differences, got: %+v", diff)
	}
}

func TestCompareEnvs_DetectsChanges(t *testing.T) {
	dataA := map[string]interface{}{"KEY": "old", "ONLY_A": "x"}
	dataB := map[string]interface{}{"KEY": "new", "ONLY_B": "y"}
	server := newMockEnvServer(t, dataA, dataB)
	defer server.Close()
	client := newEnvClient(t, server)

	diff, err := CompareEnvs(context.Background(), client, "secret/pathA", "secret/pathB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := diff.Changed["KEY"]; !ok {
		t.Error("expected KEY to be in Changed")
	}
	if _, ok := diff.OnlyInA["ONLY_A"]; !ok {
		t.Error("expected ONLY_A in OnlyInA")
	}
	if _, ok := diff.OnlyInB["ONLY_B"]; !ok {
		t.Error("expected ONLY_B in OnlyInB")
	}
}

func TestFormatEnvDiff_NoDifferences(t *testing.T) {
	diff := &EnvDiff{
		OnlyInA:   map[string]string{},
		OnlyInB:   map[string]string{},
		Changed:   map[string][2]string{},
		Identical: map[string]string{"KEY": "val"},
	}
	out := FormatEnvDiff(diff, "a", "b")
	if !strings.Contains(out, "No differences") {
		t.Errorf("expected no-diff message, got: %s", out)
	}
}

func TestFormatEnvDiff_ShowsDiff(t *testing.T) {
	diff := &EnvDiff{
		OnlyInA:   map[string]string{"A_KEY": "aval"},
		OnlyInB:   map[string]string{},
		Changed:   map[string][2]string{},
		Identical: map[string]string{},
	}
	out := FormatEnvDiff(diff, "pathA", "pathB")
	if !strings.Contains(out, "A_KEY") {
		t.Errorf("expected A_KEY in output, got: %s", out)
	}
}
