# Secure Templates - Development Guide

## Project Overview

CLI tool written in Go that renders Go templates with secret values from pluggable secret engines. Binary name: `secure-templates` (alias: `stpl`).

**Module**: `github.com/zcloud-ws/secure-templates`
**Go version**: 1.21.9
**Author**: Edimar Cardoso

## Architecture

```
main.go                     # Entry point → delegates to pkg/app
pkg/
  app/app.go                # CLI setup (urfave/cli/v2), commands: init-config, manage-secret, render
  config/config.go          # Config structs (SecureTemplateConfig, VaultConfig, LocalFileConfig)
  connectors/
    connector.go            # Connector interface (Init, Secret, WriteKey, WriteKeys, Finalize, ConnectorType)
    vault.go                # HashiCorp Vault KVv2 connector
    local-file.go           # Local file connector with RSA encryption (OAEP + SHA256)
    oci-vault.go            # OCI Vault connector (Oracle Cloud Infrastructure)
    no-connector.go         # No-op connector (used when no config is provided)
    print-keys.go           # Collects template key references (for --print-keys flag)
  envs/envs.go              # Environment variable name constants
  helpers/helpers.go         # Config parsing, RSA key generation, env file parsing
  logging/log.go             # Global logrus logger
  render/
    template.go             # Go template parsing and execution
    funcs.go                # Custom template functions: env, secret (+ sprig functions)
```

## Key Concepts

- **Connector interface**: All secret engines implement `connectors.Connector`. Factory: `connectors.NewConnector()`.
- **Secret engines**: `vault`, `local-file`, `oci-vault`, `no` (no-op), `print-keys` (introspection).
- **Template functions**: `secret "name" "key"`, `env "VAR_NAME"`, all [sprig](https://masterminds.github.io/sprig/) functions.
- **Config file**: JSON with `secret_engine`, `vault_config`, `local_file_config`, `oci_vault_config`, and `options` fields.

## Commands

| Command | Description |
|---------|-------------|
| `secure-templates init-config -o file.json` | Generate sample config with local-file engine |
| `secure-templates manage-secret put SECRET KEY VALUE` | Write a secret key-value |
| `secure-templates manage-secret import SECRET envfile` | Import secrets from .env file |
| `secure-templates [-c config.json] [-o output] template.yaml` | Render a template |
| `secure-templates -p template.yaml` | Print template key references |

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `SEC_TPL_CONFIG` | Path to config file |
| `SEC_TPL_OUTPUT` | Path to output file |
| `VAULT_ADDR` | Vault server address |
| `VAULT_TOKEN` | Vault authentication token |
| `VAULT_SECRET_ENGINE` | Vault secret engine name |
| `VAULT_NS` | Vault namespace |
| `LOCAL_SECRET_PRIVATE_KEY` | Base64-encoded RSA private key |
| `LOCAL_SECRET_PRIVATE_KEY_PASSPHRASE` | Passphrase for RSA key |
| `OCI_CONFIG_FILE` | Path to OCI config file (default: `~/.oci/config`) |
| `OCI_CONFIG_PROFILE` | OCI config profile (default: `DEFAULT`) |
| `OCI_VAULT_OCID` | OCI Vault OCID |
| `OCI_COMPARTMENT_OCID` | OCI Compartment OCID (required for write operations) |
| `OCI_KEY_OCID` | OCI Master Encryption Key OCID (required for write operations) |

## Build & Run

```bash
# Build
go build -o secure-templates .

# Run tests
go test ./...

# Run tests from test directory (tests use relative paths)
cd test && go test ./...

# Coverage
./coverage-update.sh

# Release (via GoReleaser, triggered by git tags v*)
goreleaser release --clean
```

## Testing

- Tests are in `test/` directory (separate package), not alongside source files.
- Test framework: custom `DataTest` struct with `SuiteTest()` runner in `test/test_setup.go`.
- Tests call `app.InitApp()` directly with args, capturing stdout/stderr via `bytes.Buffer`.
- Tests verify output contains expected strings (`RequiredStrings`, `RequiredErrStrings`).
- Test config and secret files are in `test/configs/`.
- Sample templates are in `test/samples/` (.env, k8s YAML, JSON).

## Development Environment

- **Dev Vault**: `dev/vault/` has Docker Compose setup with policies and init scripts.
- **IDE**: GoLand/IntelliJ (.idea/ config present).
- **CI/CD**: GitHub Actions release workflow (`.github/workflows/release.yml`) using GoReleaser + GPG signing.

## Code Conventions

- All code comments and documentation MUST be written in English.
- Package layout follows Go convention: `pkg/` for library code, `test/` for integration tests.
- Logging via `logging.Log` (global logrus instance). Use `Log.Infof`, `Log.Warnf`, `Log.Fatalf`.
- Error handling: `logging.Log.Fatalf` for unrecoverable errors, return `error` otherwise.
- Config values support environment variable expansion via `os.ExpandEnv` (any value containing `$`).
- New connectors must implement `connectors.Connector` interface and be registered in `NewConnector()` switch.
- Use `helpers.GetEnv(name, default)` to read environment variables with fallbacks.

## Dependencies

| Library | Purpose |
|---------|---------|
| `urfave/cli/v2` | CLI framework |
| `hashicorp/vault/api` | Vault API client |
| `Masterminds/sprig/v3` | Template function library |
| `joho/godotenv` | .env file parsing |
| `go-jose/go-jose/v3` | JSON serialization (used in local-file connector) |
| `sirupsen/logrus` | Structured logging |
| `oracle/oci-go-sdk/v65` | OCI SDK (Vault, Secrets clients) |