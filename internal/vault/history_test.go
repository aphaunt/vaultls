package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockVersionServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/myapp/config":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"versions": map[string]interface{}{
						"1": map[string]interface{}{"destroyed": false},
						"2": map[string]interface{}{"destroyed": false},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestDiffVersions_SameData(t *testing.T) {
	dataA := map[string]string{"key": "value"}
	dataB := map[string]string{"key": "value"}
	results := DiffSecrets(dataA, dataB)
	assert.Empty(t, results)
}

func TestDiffVersions_ChangedKey(t *testing.T) {
	dataA := map[string]string{"key": "old"}
	dataB := map[string]string{"key": "new"}
	results := DiffSecrets(dataA, dataB)
	require.Len(t, results, 1)
	assert.Equal(t, "key", results[0].Key)
	assert.Equal(t, "old", results[0].ValueA)
	assert.Equal(t, "new", results[0].ValueB)
}

func TestDiffVersions_AddedKey(t *testing.T) {
	dataA := map[string]string{}
	dataB := map[string]string{"newkey": "val"}
	results := DiffSecrets(dataA, dataB)
	require.Len(t, results, 1)
	assert.Equal(t, "newkey", results[0].Key)
	assert.Equal(t, "", results[0].ValueA)
	assert.Equal(t, "val", results[0].ValueB)
}

func TestGetVersion_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	server := newMockVersionServer(t)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token", "secret")
	require.NoError(t, err)

	_, err = client.GetVersion(ctx, "myapp/config", 1)
	assert.Error(t, err)
}
