package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockValidateServer(data map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]interface{}{
			"data": map[string]interface{}{"data": data},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
}

func TestValidateSecrets_AllValid(t *testing.T) {
	srv := newMockValidateServer(map[string]interface{}{"DB_HOST": "localhost", "DB_PORT": "5432"})
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-token")
	result, err := ValidateSecrets(context.Background(), client, "secret/data/app", []string{"DB_HOST", "DB_PORT"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsValid() {
		t.Errorf("expected valid result, got missing=%v empty=%v", result.Missing, result.Empty)
	}
	if len(result.Valid) != 2 {
		t.Errorf("expected 2 valid keys, got %d", len(result.Valid))
	}
}

func TestValidateSecrets_MissingKey(t *testing.T) {
	srv := newMockValidateServer(map[string]interface{}{"DB_HOST": "localhost"})
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-token")
	result, err := ValidateSecrets(context.Background(), client, "secret/data/app", []string{"DB_HOST", "DB_PORT"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsValid() {
		t.Error("expected invalid result due to missing key")
	}
	if len(result.Missing) != 1 || result.Missing[0] != "DB_PORT" {
		t.Errorf("expected DB_PORT in missing, got %v", result.Missing)
	}
}

func TestValidateSecrets_EmptyValue(t *testing.T) {
	srv := newMockValidateServer(map[string]interface{}{"DB_HOST": "localhost", "DB_PORT": ""})
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-token")
	result, err := ValidateSecrets(context.Background(), client, "secret/data/app", []string{"DB_HOST", "DB_PORT"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsValid() {
		t.Error("expected invalid result due to empty value")
	}
	if len(result.Empty) != 1 || result.Empty[0] != "DB_PORT" {
		t.Errorf("expected DB_PORT in empty, got %v", result.Empty)
	}
}
