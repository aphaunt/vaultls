package vault_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lockStore struct {
	mu   sync.Mutex
	data map[string]map[string]interface{}
}

func newStatefulLockServer(t *testing.T) (*httptest.Server, *lockStore) {
	t.Helper()
	ls := &lockStore{data: map[string]map[string]interface{}{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ls.mu.Lock()
		defer ls.mu.Unlock()
		path := r.URL.Path
		switch r.Method {
		case http.MethodGet:
			data, ok := ls.data[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
		case http.MethodPost:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			ls.data[path] = body
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			delete(ls.data, path)
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(server.Close)
	return server, ls
}

func TestLockIntegration_LockAndUnlock(t *testing.T) {
	server, _ := newStatefulLockServer(t)

	client, err := NewClient(server.URL, "test-token")
	require.NoError(t, err)

	path := "secret/data/env/prod"

	err = LockSecret(client, path, "alice", "release freeze")
	require.NoError(t, err)

	lock, err := GetLock(client, path)
	require.NoError(t, err)
	require.NotNil(t, lock)
	assert.Equal(t, "alice", lock.LockedBy)
	assert.Equal(t, "release freeze", lock.Reason)

	err = UnlockSecret(client, path, "alice")
	require.NoError(t, err)

	lock, err = GetLock(client, path)
	require.NoError(t, err)
	assert.Nil(t, lock)
}

func TestLockIntegration_UnauthorizedUnlock(t *testing.T) {
	server, _ := newStatefulLockServer(t)

	client, err := NewClient(server.URL, "test-token")
	require.NoError(t, err)

	path := "secret/data/env/staging"

	err = LockSecret(client, path, "alice", "testing")
	require.NoError(t, err)

	err = UnlockSecret(client, path, "bob")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not authorized")
}
