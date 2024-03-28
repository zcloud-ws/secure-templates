# Configure Vault with MFA TOTP and userpass auth

This document describes how to configure Vault with Multi-Factor Authentication (MFA) using Time-based One-Time Passwords (TOTP) alongside the userpass authentication method. 

## Prerequisites

* You have administrative access to a Vault instance.
* You have downloaded and sourced the helper script `func-utils.sh`. (You might include a link to this script or instructions on creating your own).

## Administrator Setup

Start a Vault CLI Container

* `VAULT_ADDR`: Vault API URL. Ex.: `http://localhost:8200`
* `VAULT_TOKEN`: Admin token with permission manage Vault resources. 

```shell
docker run --rm -it --name vault-cli \
  -e VAULT_ADDR="${VAULT_ADDR}" \
  -e VAULT_TOKEN="${VAULT_TOKEN}" \
  hashicorp/vault:1.15 sh
```

### Load function helpers

```shell
. <(wget -q -O- https://raw.githubusercontent.com/edimarlnx/secure-templates/main/dev/vault/mfa-with-username/func-utils.sh)
```

**Security Note:** Review the contents of [`user-func-utils.sh`](https://raw.githubusercontent.com/edimarlnx/secure-templates/main/dev/vault/mfa-with-username/func-utils.sh) for transparency.


### Enable the userpass Auth Method

```shell
enable_auth_userpass
```

### Enable the KV Secrets Engine

Enable KV with name `staging`

```shell
enable_kv2 staging
```

### Create the auth_totp ACL Policy

This policy grants users the ability to manage their TOTP secrets and change their passwords.

```shell
create_auth_acl
```

### Create the staging ACL Policy

This policy allows read, write, and delete access to secrets within the `staging/data/staging/*` path.

```shell
create_acl staging "staging/data/staging/*" read,write,delete
```

### Enable the TOTP MFA Method

Enable the TOTP engine with name `totp` and issuer name `MyOrg`

```shell
create_mfa_totp totp MyOrg
```

### Create a User with KV Access

* Create a user named 'john'.
* Generate a secure, random password.
* Assign the `staging_read` and `default` policies.

```shell
USERPWD="$(date | md5 | cut -b-20)"
create_user john "$USERPWD" "staging_read,default"
```

### Generate a Temporary User Token

* Allow 'john' to set their password and configure TOTP.
* Set an appropriate token expiration (12h in this example).

```shell
#USERPWD random password generated in prior step. 
generate_user_token john "$USERPWD" 12h
```

### Enforce MFA for the User

Require 'john' to use TOTP during login.

```shell
create_login_enforcement_entity totp john
```


## User Setup

Start a Vault CLI Container

```shell
docker run --rm -it --name vault-cli \
  -e VAULT_ADDR="${VAULT_ADDR}" \
  hashicorp/vault:1.15 sh
```

Update Password and Generate TOTP Secret

```shell
. <(wget -q -O- https://raw.githubusercontent.com/edimarlnx/secure-templates/main/dev/vault/mfa-with-username/user-func-utils.sh)
user_update_password USER_TOKEN USERNAME NEW_PASSWORD
user_generate_totp_secret USER_TOKEN METHOD_NAME USERNAME
```

**Security Note:** Review the contents of [`user-func-utils.sh`](https://raw.githubusercontent.com/edimarlnx/secure-templates/main/dev/vault/mfa-with-username/user-func-utils.sh) for transparency.

## Additional Administrator Tasks (Optional)

### Disable MFA for a User (if needed)

```shell
create_login_enforcement_entity totp john delete
```

### Remove a User's TOTP Secret (if needed)

```shell
destroy_totp_secret totp john
```
