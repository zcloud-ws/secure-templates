# Secure Templates

[![Coverage](https://img.shields.io/badge/Coverage-64%25-orange)](coverage/cover.html)

A CLI tool that renders [Go templates](https://pkg.go.dev/text/template) with secret values loaded from pluggable secret engines. Keep your templates in version control and inject secrets at render time.

**Binary:** `secure-templates` (alias: `stpl`)

## Features

- Render any text file using Go template syntax with secret injection
- Pluggable secret engines: HashiCorp Vault, local encrypted file, or no-op
- Built-in [sprig](https://masterminds.github.io/sprig/) template functions (100+ utility functions)
- Custom template delimiters to avoid conflicts with Helm, Jinja, etc.
- Environment variable support in templates and config values
- Secret introspection with `--print-keys`

## Installation

### Linux and macOS

Install to `/usr/local/bin` (requires sudo):

```shell
curl -s https://raw.githubusercontent.com/quave-one/secure-templates/main/scripts/install.sh | sudo sh -
```

Install to current directory (no sudo):

```shell
curl -s https://raw.githubusercontent.com/quave-one/secure-templates/main/scripts/install.sh | sh -
```

### Manual / Windows

Download the binary for your platform from the [releases page](https://github.com/quave-one/secure-templates/releases).

## Quick Start

```shell
# 1. Generate a config with local-file secret engine
secure-templates init-config -o config.json

# 2. Store a secret
secure-templates -c config.json manage-secret put myapp db_password "s3cr3t"

# 3. Create a template
echo 'DB_PASSWORD={{ secret "myapp" "db_password" }}' > app.env.tpl

# 4. Render
secure-templates -c config.json app.env.tpl
# Output: DB_PASSWORD=s3cr3t
```

## Template Functions

### `secret`

Retrieves a value from the configured secret engine.

```
{{ secret "secret_name" "key_name" }}
```

When called with only the secret name, returns a key-value map that can be iterated:

```
{{ range $key, $value := secret "secret_name" -}}
{{ $key }}={{ $value }}
{{ end }}
```

### `env`

Reads an environment variable:

```
{{ env "MY_VAR" }}
```

### Sprig functions

All [sprig](https://masterminds.github.io/sprig/) functions are available. Common examples:

```
{{ secret "app" "password" | b64enc }}      # base64 encode
{{ env "HOST" | upper }}                     # uppercase
{{ secret "app" "name" | default "myapp" }}  # default value
```

## Commands

### `init-config`

Generate a sample config file with the `local-file` secret engine:

```shell
secure-templates init-config -o config.json
```

| Flag | Env var | Description |
|------|---------|-------------|
| `--output`, `-o` | `SEC_TPL_OUTPUT` | Output file path (stdout if omitted) |
| `--secret-file` | | Path for the secret data file |
| `--private-key-passphrase` | `LOCAL_SECRET_PRIVATE_KEY_PASSPHRASE` | Passphrase for RSA key encryption |

### `manage-secret`

Manage secrets in the configured engine.

**Add or update a single key:**

```shell
secure-templates -c config.json manage-secret put <SECRET> <KEY> <VALUE>
```

**Import keys from an .env file:**

```shell
secure-templates -c config.json manage-secret import <SECRET> <ENV_FILE>
```

### Render (default action)

Render a template file using values from the configured secret engine:

```shell
secure-templates -c config.json [flags] <TEMPLATE_FILE>
```

| Flag | Env var | Description |
|------|---------|-------------|
| `--config`, `-c` | `SEC_TPL_CONFIG` | Path to config file |
| `--output`, `-o` | `SEC_TPL_OUTPUT` | Output file (stdout if omitted) |
| `--print-keys`, `-p` | | List secret key references used in the template |
| `--left-delim`, `--ld` | `SEC_TPL_LEFT_DELIM` | Custom left template delimiter |
| `--right-delim`, `--rd` | `SEC_TPL_RIGHT_DELIM` | Custom right template delimiter |

## Custom Template Delimiters

When rendering templates that target systems using Go template syntax (e.g., Helm charts), the default `{{ }}` delimiters conflict. Use custom delimiters so that `secure-templates` only processes its own expressions while standard `{{ }}` passes through untouched.

**Example** - a Helm values file using `<< >>` delimiters:

Template (`values.yaml.tpl`):
```yaml
app_user: {{ .Values.appUser }}
app_password: << secret "core" "app_passwd" >>
```

Render:
```shell
secure-templates --left-delim "<<" --right-delim ">>" values.yaml.tpl
```

Output:
```yaml
app_user: {{ .Values.appUser }}
app_password: s3cr3t_v4lu3
```

Custom delimiters can also be set in the config file:

```json
{
  "options": {
    "leftDelim": "<<",
    "rightDelim": ">>"
  }
}
```

CLI flags take precedence over config file values.

## Supported Secret Engines

### HashiCorp Vault

Uses the [Vault KVv2](https://www.vaultproject.io/) secret engine.

| Env var | Description |
|---------|-------------|
| `VAULT_ADDR` | Vault server address |
| `VAULT_TOKEN` | Authentication token |
| `VAULT_SECRET_ENGINE` | Secret engine name |
| `VAULT_NS` | Vault namespace |

For local development, a Docker Compose setup is available in [`dev/vault/`](dev/vault/README.md).

### Local File

Stores secrets in a local JSON file encrypted with RSA (OAEP + SHA256).

| Env var | Description |
|---------|-------------|
| `LOCAL_SECRET_PRIVATE_KEY` | Base64-encoded RSA private key |
| `LOCAL_SECRET_PRIVATE_KEY_PASSPHRASE` | Passphrase for the RSA key |

## Config File

```json
{
  "secret_engine": "local-file",
  "vault_config": {
    "address": "http://localhost:8200",
    "token": "token",
    "secret_engine": "kv",
    "ns": "dev"
  },
  "local_file_config": {
    "filename": "secrets.json",
    "enc_priv_key": "LS0tLS...."
  },
  "options": {
    "secretShowNameAsValueIfEmpty": false,
    "secretIgnoreNotFoundKey": false,
    "envShowNameAsValueIfEmpty": false,
    "envAllowAccessToSecureTemplateEnvs": false,
    "envRestrictedNameRegex": "SC_.+",
    "leftDelim": "",
    "rightDelim": ""
  }
}
```

Config values support environment variable expansion: any value containing `$` is expanded (e.g., `"$VAULT_TOKEN"`).

### Options Reference

| Option | Default | Description |
|--------|---------|-------------|
| `secretShowNameAsValueIfEmpty` | `false` | Show the key name as value when the secret value is empty |
| `secretIgnoreNotFoundKey` | `false` | Ignore missing keys instead of failing |
| `envShowNameAsValueIfEmpty` | `false` | Show the variable name as value when the env var is empty |
| `envAllowAccessToSecureTemplateEnvs` | `false` | Allow `env` function to access `secure-templates` internal env vars |
| `envRestrictedNameRegex` | `""` | Regex pattern for restricted env var names (e.g., `SC_.+`) |
| `leftDelim` | `""` | Custom left template delimiter (empty = `{{`) |
| `rightDelim` | `""` | Custom right template delimiter (empty = `}}`) |

## Examples

### .env file

Template ([source](./test/samples/.env)):
```
export APP_USER={{ secret "core" "app_user" }}
export APP_PASSWORD={{ secret "core" "app_passwd" }}
```

Output:
```
export APP_USER=dev_user
export APP_PASSWORD=2dabe3d7c66fb75f751202fdab19266b
```

### Kubernetes Secret

Template ([source](./test/samples/k8s-secret.yaml)):
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {{ env "SECRET_NAME" }}
  namespace: {{ env "SECRET_NAMESPACE" }}
data:
  APP_USER: {{ secret "core" "app_user" | b64enc }}
  APP_PASSWORD: {{ secret "core" "app_passwd" | b64enc }}
stringData:
  CLIENT_APP_USER: "{{ secret "client" "app_user" }}"
  CLIENT_APP_PASSWORD: "{{ secret "client" "app_passwd" }}"
```

### Iterating over secret keys

Template ([source](./test/samples/secrets-list.env)):
```
{{ range $key, $value := secret "test" -}}
{{ $key }}:{{ $value }}
{{ end }}
```

## Building from Source

```bash
# Build
go build -o secure-templates .

# Run tests
cd test && go test ./...

# Coverage report
./coverage-update.sh
```

Requires Go 1.21+.

## Author

Edimar Cardoso

- [edimarlnx@gmail.com](mailto:edimarlnx@gmail.com)
- [edimar@quave.one](mailto:edimar@quave.one)
- [oss@quave.one](mailto:oss@quave.one)
- Website: [www.quave.one](https://www.quave.one)

## License

[MIT](./LICENSE)
