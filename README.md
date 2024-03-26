# Secure Templates

Secure Templates is a tool to render templates using go-templates and load data values from secrets engine.

## Installation for Linux and Mac

### Install on /usr/local/bin with sudo

```shell
curl -s https://raw.githubusercontent.com/edimarlnx/secure-templates/main/scripts/install.sh | sudo sh -
```

### Install on work directory without sudo

```shell
curl -s https://raw.githubusercontent.com/edimarlnx/secure-templates/main/scripts/install.sh | sh -
```

## Manual installation and Windows users

Go to [releases page](https://github.com/edimarlnx/secure-templates/releases) and download the file according your
system.

## Samples

- [.env](./test/samples/.env): [.rendered-env](./test/samples/.rendered-env) 
- [k8s-secret.yaml](./test/samples/k8s-secret.yaml): [k8s-secret-rendered.yaml](./test/samples/k8s-secret-rendered.yaml)

## Supported Secrets engines

- [Vault](https://www.vaultproject.io/): A free Vault solution by [HashiCorp](https://www.hashicorp.com/)
    - In development, you can use a docker to run Vault. [See](dev/vault/README.md)
- Local file: Local file using rsa key par to encrypt data

## Config file

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
    "filename": "test/local-file-secret.json",
    "enc_priv_key": "LS0tLS...."
  }
}
```

## Commands

### init-config

Initialize a sample config with `local-file` as a secret engine

```shell
secure-templates init-config -o local-file-cfg.json
```

#### Options

```text
NAME:
   secure-templates init-config - Init a sample config

USAGE:
   secure-templates init-config [command options] [arguments...]

OPTIONS:
   --output value, -o value, --out value   [$SEC_TPL_OUTPUT]
   --secret-file value                    (default: "./test/local-file-secret.json")
   --private-key-passphrase value         [$LOCAL_SECRET_PRIVATE_KEY_PASSPHRASE]
   --help, -h                             show help
2024/03/03 00:46:34 ERROR Required flag "config" not set
```

#### Environment variables

- `SEC_TPL_OUTPUT`: Path to output config file.
- `LOCAL_SECRET_PRIVATE_KEY_PASSPHRASE`: Passphrase to encrypt private key.

### manage-secret

Manage secret engine

Create or update the key `app_passwd` into secret `core` with value `abc123`

```shell
secure-templates manage-secret put core app_passwd abc123
```

#### Subcommands arguments

- `put`: `SECRET` `KEY` `VALUE`
- `import`: `SECRET` `ENV FILE`

#### Options

```text
NAME:
   secure-templates manage-secret - Manage secret

USAGE:
   secure-templates manage-secret command [command options] 

COMMANDS:
   put      Add or update key value
   import   Add or update key value using env file
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --config value, -c value, --cfg value   [$SEC_TPL_CONFIG]
   --help, -h                             show help
```

#### Environment variables

- `LOCAL_SECRET_PRIVATE_KEY`: Private key encoded with base64.
- `LOCAL_SECRET_PRIVATE_KEY_PASSPHRASE`: Passphrase to decrypt private key.

### Template Render

Render template using values from configured secret engine

Render a template file

```shell
secure-templates FILEPATH
```

#### Arguments

- `FILEPATH`: Filepath for template to render.

#### Options

```text
NAME:
   secure-templates - A template render tool

USAGE:
   secure-templates [global options] command [command options] 

VERSION:
   dev

DESCRIPTION:
   Secure Templates is a tool to render templates using go-templates and load data values from secrets engine.

COMMANDS:
   init-config    Init a sample config
   manage-secret  Manage secret
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value, --cfg value   [$SEC_TPL_CONFIG]
   --output value, -o value, --out value   [$SEC_TPL_OUTPUT]
   --print-keys, -p                       (default: false)
   --help, -h                             show help
   --version, -v                          print the version
```

#### Environment variables

- `SEC_TPL_CONFIG`: Path to config file.
- `SEC_TPL_OUTPUT`: Path to output template file.
- `VAULT_TOKEN`: Vault token to call the Vault API.
- `LOCAL_SECRET_PRIVATE_KEY`: Private key encoded with base64. Used only for `local-secret` engine.
- `LOCAL_SECRET_PRIVATE_KEY_PASSPHRASE`: Passphrase to decrypt private key. Used only for `local-secret` engine.

## Template Functions

* `base64Encode`: Encode a base 64 string,
* `base64Decode`: Decode a base 64 string,
* `env`: Get environment variable,
* `secret`: Get the Key value of a secret engine. If the key name is not provided, it returns a key and value map that can be iterated.
* `toUpper`: Convert string to upper case
* `toLower`: Convert string to lower case
* `trimSpace`: Trim string spaces

# Author

Edimar Cardoso

Emails: [edimarlnx@gmail.com](mailto:edimarlnx@gmail.com) [edimar@zcloud.ws](mailto:edimar@zcloud.ws)

Website: www.zcloud.ws

# License

[MIT](./LICENSE)