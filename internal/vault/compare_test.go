package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockCompareServer(dataA, dataB map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		switch r.URL.Path {
		case "/v1/secret/data/pathA":
			data = dataA
		case "/v1/secret/data/pathB":
			data = dataB
		default:
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": data},
		})
	}))
}

func TestCompareSecrets_Identical(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	server := newMockCompareServer(data, data)
	defer server.Close()

	client, _ := NewClient(server.URL, "test-token")
	result, err := CompareSecrets(context.Background(), client, "secret/pathA", "secret/pathB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Identical) != 1 || result.Identical[0] != "key" {
		t.Errorf("expected identical key 'key', got %v", result.Identical)
	}
	if len(result.Changed) != 0 || len(result.OnlyInA) != 0 || len(result.OnlyInB) != 0 {
		t.Errorf("expected no differences")
	}
}

func TestCompareSecrets_Changed(t *testing.T) {
	dataA := map[string]interface{}{"key": "valueA", "shared": "same"}
	dataB := map[string]interface{}{"key": "valueB", "shared": "same"}
	server := newMockCompareServer(dataA, dataB)
	defer server.Close()

	client, _ := NewClient(server.URL, "test-token")
	result, err := CompareSecrets(context.Background(), client, "secret/pathA", "secret/pathB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pair, ok := result.Changed["key"]; !ok || pair[0] != "valueA" || pair[1] != "valueB" {
		t.Errorf("expected changed key 'key', got %v", result.Changed)
	}
	if len(result.Identical) != 1 || result.Identical[0] != "shared" {
		t.Errorf("expected identical 'shared', got %v", result.Identical)
	}
}

func TestCompareSecrets_DisjointKeys(t *testing.T) {
	dataA := map[string]interface{}{"onlyA": "val"}
	dataB := map[string]interface{}{"onlyB": "val"}
	server := newMockCompareServer(dataA, dataB)
	defer server.Close()

	client, _ := NewClient(server.URL, "test-token")
	result, err := CompareSecrets(context.Background(), client, "secret/pathA", "secret/pathB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.OnlyInA) != 1 || result.OnlyInA[0] != "onlyA" {
		t.Errorf("expected onlyA in OnlyInA, got %v", result.OnlyInA)
	}
	if len(result.OnlyInB) != 1 || result.OnlyInB[0] != "onlyB" {
		t.Errorf("expected onlyB in OnlyInB, got %v", result.OnlyInB)
	}
}
