package vault_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/vaultls/internal/vault"
)

func newMockVaultServer(t *testing.T, path string, responseBody string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	}))
}

func TestNewClient_MissingAddress(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")

	_, err := vault.NewClient(vault.Config{})
	if err == nil {
		t.Fatal("expected error when address is missing, got nil")
	}
}

func TestNewClient_MissingToken(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")

	_, err := vault.NewClient(vault.Config{Address: "http://127.0.0.1:8200"})
	if err == nil {
		t.Fatal("expected error when token is missing, got nil")
	}
}

func TestNewClient_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := vault.NewClient(vault.Config{
		Address: server.URL,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestList_ReturnsKeys(t *testing.T) {
	body := `{"data":{"keys":["foo","bar","baz/"]}}`
	server := newMockVaultServer(t, "/v1/secret/metadata/test", body)
	defer server.Close()

	client, err := vault.NewClient(vault.Config{
		Address: server.URL,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	keys, err := client.List("secret/metadata/test")
	if err != nil {
		t.Fatalf("unexpected error listing: %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
}
