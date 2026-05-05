package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newMockBookmarkServer(t *testing.T) (*httptest.Server, *api.Client) {
	t.Helper()

	var mu sync.Mutex
	store := map[string]interface{}{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			if val, ok := store[r.URL.Path]; ok {
				json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": val}})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			store[r.URL.Path] = body["data"]
			w.WriteHeader(http.StatusNoContent)
		}
	})

	ts := httptest.NewServer(mux)
	cfg := api.DefaultConfig()
	cfg.Address = ts.URL
	client, _ := api.NewClient(cfg)
	client.SetToken("test-token")
	return ts, client
}

func TestAddBookmark_EmptyName(t *testing.T) {
	_, client := newMockBookmarkServer(t)
	err := AddBookmark(context.Background(), client, "", "secret/foo", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestAddBookmark_EmptyPath(t *testing.T) {
	_, client := newMockBookmarkServer(t)
	err := AddBookmark(context.Background(), client, "mykey", "", "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestRemoveBookmark_NotFound(t *testing.T) {
	_, client := newMockBookmarkServer(t)
	err := RemoveBookmark(context.Background(), client, "ghost")
	if err == nil {
		t.Fatal("expected error for missing bookmark")
	}
}

func TestListBookmarks_Empty(t *testing.T) {
	_, client := newMockBookmarkServer(t)
	list, err := ListBookmarks(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %d", len(list))
	}
}
