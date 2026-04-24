package vault

import (
	"fmt"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client.
type Client struct {
	api *vaultapi.Client
}

// Config holds configuration for creating a Vault client.
type Config struct {
	Address string
	Token   string
}

// NewClient creates a new Vault client using the provided config.
// Falls back to environment variables VAULT_ADDR and VAULT_TOKEN.
func NewClient(cfg Config) (*Client, error) {
	defaultCfg := vaultapi.DefaultConfig()

	addr := cfg.Address
	if addr == "" {
		addr = os.Getenv("VAULT_ADDR")
	}
	if addr == "" {
		return nil, fmt.Errorf("vault address not set: use --addr flag or VAULT_ADDR env var")
	}
	defaultCfg.Address = addr

	client, err := vaultapi.NewClient(defaultCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	token := cfg.Token
	if token == "" {
		token = os.Getenv("VAULT_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("vault token not set: use --token flag or VAULT_TOKEN env var")
	}
	client.SetToken(token)

	return &Client{api: client}, nil
}

// List returns the keys at the given Vault path.
func (c *Client) List(path string) ([]string, error) {
	secret, err := c.api.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list path %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	raw, ok := secret.Data["keys"]
	if !ok {
		return nil, fmt.Errorf("unexpected response: missing 'keys' field")
	}

	ifaces, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for 'keys' field")
	}

	keys := make([]string, 0, len(ifaces))
	for _, v := range ifaces {
		if s, ok := v.(string); ok {
			keys = append(keys, s)
		}
	}
	return keys, nil
}

// Read returns the secret data at the given Vault path.
func (c *Client) Read(path string) (map[string]interface{}, error) {
	secret, err := c.api.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read path %q: %w", path, err)
	}
	if secret == nil {
		return nil, nil
	}
	return secret.Data, nil
}
