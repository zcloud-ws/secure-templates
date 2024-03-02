# Secure Templates

Secure Templates is a tool to render templates using go-templates and load data values from secrets engine.

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
   --help, -h                             show help
```

#### Environment variables

- `SEC_TPL_OUTPUT`: Path to output config file.

### local-secret

Manage local file secret engine

Create or update the key `app_passwd` into secret `core` with value `abc123` 
```shell
secure-templates local-secret put core app_passwd abc123
```
#### Subcommands arguments

- `put`: `SECRET` `KEY` `VALUE`  

#### Options

```text
NAME:
   secure-templates local-secret - Manipulate local secret file

USAGE:
   secure-templates local-secret command [command options] 

COMMANDS:
   put      Add or update key value
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --config value, -c value, --cfg value  (default: "test/local-file-cfg.json") [$SEC_TPL_CONFIG]
   --help, -h                             show help
```

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
   init-config   Init empty config
   local-secret  Manage local secret file
   help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value, --cfg value   [$SEC_TPL_CONFIG]
   --output value, -o value, --out value   [$SEC_TPL_OUTPUT]
   --help, -h                             show help
   --version, -v                          print the version
```

#### Environment variables

- `SEC_TPL_CONFIG`: Path to config file.
- `SEC_TPL_OUTPUT`: Path to output template file.


## Template Functions

* `base64Encode`: Encode a base 64 string,
* `base64Decode`: Decode a base 64 string,
* `env`: Get environment variable,
* `secret`: Get key value from a secret engine
* `toUpper`: Convert string to upper case
* `toLower`: Convert string to lower case
* `trimSpace`: Trim string spaces

# Author

Edimar Cardoso 

Emails: [edimarlnx@gmail.com](mailto:edimarlnx@gmail.com) [edimar@zcloud.ws](mailto:edimar@zcloud.ws)

Website: www.zcloud.ws

# License

[MIT](./LICENSE)