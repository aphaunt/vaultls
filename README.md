# vaultls

A CLI for browsing and diffing HashiCorp Vault secrets across environments.

---

## Installation

```bash
go install github.com/yourusername/vaultls@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultls.git && cd vaultls && go build -o vaultls .
```

---

## Usage

Browse secrets in a Vault path:

```bash
vaultls ls secret/data/myapp
```

Diff secrets between two environments:

```bash
vaultls diff secret/data/myapp/staging secret/data/myapp/production
```

Export a path as a tree:

```bash
vaultls ls --tree secret/data/myapp
```

> **Note:** Ensure `VAULT_ADDR` and `VAULT_TOKEN` environment variables are set before use.

```bash
export VAULT_ADDR=https://vault.example.com
export VAULT_TOKEN=s.yourtoken
```

---

## Requirements

- Go 1.21+
- HashiCorp Vault with a valid token and appropriate policies

---

## License

MIT © 2024 [yourusername](https://github.com/yourusername)