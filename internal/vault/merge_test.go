package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func newMockMergeServer(t *testing.T) (*httptest.Server, *sync.Map) {
	t.Helper()
	store := &sync.Map{}

	store.Store("/v1/secret/data/src", map[string]interface{}{"a": "1", "b": "2"})
	store.Store("/v1/secret/data/dst", map[string]interface{}{"b": "old", "c": "3"})

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			val, ok := store.Load(r.URL.Path)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": val}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if data, ok := body["data"]; ok {
				store.Store(r.URL.Path, data)
			}
			w.WriteHeader(http.StatusOK)
		}
	})
	return httptest.NewServer(mux), store
}

func TestMergeSecrets_EmptyPaths(t *testing.T) {
	client := &Client{Address: "http://localhost", Token: "tok"}
	_, err := MergeSecrets(context.Background(), client, "", "dst", false, false)
	if err == nil {
		t.Fatal("expected error for empty src")
	}
}

func TestMergeSecrets_SamePath(t *testing.T) {
	client := &Client{Address: "http://localhost", Token: "tok"}
	_, err := MergeSecrets(context.Background(), client, "secret/data/x", "secret/data/x", false, false)
	if err == nil {
		t.Fatal("expected error for same src and dst")
	}
}

func TestMergeSecrets_NoOverwrite(t *testing.T) {
	srv, _ := newMockMergeServer(t)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "token")
	result, err := MergeSecrets(context.Background(), client, "secret/data/src", "secret/data/dst", false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Written) != 1 || result.Written[0] != "a" {
		t.Errorf("expected Written=[a], got %v", result.Written)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "b" {
		t.Errorf("expected Skipped=[b], got %v", result.Skipped)
	}
}

func TestMergeSecrets_WithOverwrite(t *testing.T) {
	srv, _ := newMockMergeServer(t)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "token")
	result, err := MergeSecrets(context.Background(), client, "secret/data/src", "secret/data/dst", true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Overwritten) != 1 || result.Overwritten[0] != "b" {
		t.Errorf("expected Overwritten=[b], got %v", result.Overwritten)
	}
	if len(result.Written) != 1 || result.Written[0] != "a" {
		t.Errorf("expected Written=[a], got %v", result.Written)
	}
}
