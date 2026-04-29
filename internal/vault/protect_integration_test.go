package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newStatefulProtectServer(t *testing.T) (*httptest.Server, *sync.Map) {
	t.Helper()
	store := &sync.Map{}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch r.Method {
		case http.MethodGet:
			val, ok := store.Load(path)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": val})
		case http.MethodPut, http.MethodPost:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			store.Store(path, body)
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	return svr, store
}

func TestProtectIntegration_ProtectAndCheck(t *testing.T) {
	svr, store := newStatefulProtectServer(t)
	defer svr.Close()

	store.Store("secret/app", map[string]interface{}{"DB_PASS": "s3cr3t"})

	cfg := vaultapi.DefaultConfig()
	cfg.Address = svr.URL
	client, _ := vaultapi.NewClient(cfg)
	client.SetToken("test-token")

	ctx := context.Background()

	if err := ProtectSecrets(ctx, client, "secret/app"); err != nil {
		t.Fatalf("ProtectSecrets failed: %v", err)
	}

	protected, err := IsProtected(ctx, client, "secret/app")
	if err != nil {
		t.Fatalf("IsProtected failed: %v", err)
	}
	if !protected {
		t.Error("expected secret to be protected")
	}

	if err := UnprotectSecrets(ctx, client, "secret/app"); err != nil {
		t.Fatalf("UnprotectSecrets failed: %v", err)
	}

	protected, err = IsProtected(ctx, client, "secret/app")
	if err != nil {
		t.Fatalf("IsProtected after unprotect failed: %v", err)
	}
	if protected {
		t.Error("expected secret to be unprotected after removal")
	}

	val, _ := store.Load("secret/app")
	data := val.(map[string]interface{})
	if data["DB_PASS"] != "s3cr3t" {
		t.Errorf("original data should be preserved, got: %v", data)
	}
}
