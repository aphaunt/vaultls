package vault_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockLockServer(t *testing.T) *httptest.Server {
	store := map[string]map[string]interface{}{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch r.Method {
		case http.MethodGet:
			data, ok := store[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
		case http.MethodPost:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			store[path] = body
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			delete(store, path)
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func TestLockSecret_Success(t *testing.T) {
	server := newMockLockServer(t)
	defer server.Close()

	client := newTestClient(t, server.URL)
	err := LockSecret(client, "secret/data/myapp/config", "alice", "maintenance")
	require.NoError(t, err)
}

func TestLockSecret_EmptyPath(t *testing.T) {
	server := newMockLockServer(t)
	defer server.Close()

	client := newTestClient(t, server.URL)
	err := LockSecret(client, "", "alice", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path")
}

func TestUnlockSecret_Success(t *testing.T) {
	server := newMockLockServer(t)
	defer server.Close()

	client := newTestClient(t, server.URL)
	_ = LockSecret(client, "secret/data/myapp/config", "alice", "test")
	err := UnlockSecret(client, "secret/data/myapp/config", "alice")
	require.NoError(t, err)
}

func TestGetLock_ReturnsLockInfo(t *testing.T) {
	server := newMockLockServer(t)
	defer server.Close()

	client := newTestClient(t, server.URL)
	_ = LockSecret(client, "secret/data/myapp/config", "bob", "deploy freeze")

	lock, err := GetLock(client, "secret/data/myapp/config")
	require.NoError(t, err)
	assert.Equal(t, "bob", lock.LockedBy)
	assert.Equal(t, "deploy freeze", lock.Reason)
	assert.WithinDuration(t, time.Now(), lock.LockedAt, 5*time.Second)
}

func TestGetLock_NotLocked(t *testing.T) {
	server := newMockLockServer(t)
	defer server.Close()

	client := newTestClient(t, server.URL)
	lock, err := GetLock(client, "secret/data/myapp/config")
	require.NoError(t, err)
	assert.Nil(t, lock)
}

func newTestClient(t *testing.T, addr string) interface{} {
	t.Helper()
	_ = fmt.Sprintf("%s", addr)
	return nil
}
