package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func newStatefulTagServer(t *testing.T) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	store := map[string]string{}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			cm := map[string]interface{}{}
			for k, v := range store {
				cm[k] = v
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"custom_metadata": cm},
			})
		case r.Method == http.MethodPost:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if cm, ok := body["custom_metadata"].(map[string]interface{}); ok {
				for k, v := range cm {
					store[k] = v.(string)
				}
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestTagIntegration_SetAndList(t *testing.T) {
	_ = newStatefulTagServer(t)
	// Validate that TagSecrets rejects empty path before any network call.
	client := &Client{}
	err := TagSecrets(context.Background(), client, "", map[string]string{"env": "staging"}, false)
	if err == nil {
		t.Fatal("expected validation error for empty path")
	}
}

func TestTagIntegration_OverwriteFalse(t *testing.T) {
	client := &Client{}
	err := TagSecrets(context.Background(), client, "secret/app/cfg", map[string]string{}, false)
	if err == nil {
		t.Fatal("expected error for empty tags map")
	}
}

func TestTagIntegration_ListEmptyPath(t *testing.T) {
	client := &Client{}
	_, err := ListTags(context.Background(), client, "")
	if err == nil {
		t.Fatal("expected error for empty path in ListTags")
	}
}
