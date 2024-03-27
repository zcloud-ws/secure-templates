# Configure Vault with MFA TOTP and userpass auth

This document describe how to configure vault with MFA using TOTP and userpass.

## Functions definitions

The following steps considering that you had load the functions from file `func-utils.sh`

```shell
. func-utils.sh
```

## As a admin user

### Enabling userpass auth

Enable `userpass` auth method

```shell
enable_auth_userpass
```

### Enabling key value engine

Enable KV with name `staging`

```shell
enable_kv2 staging
```

### Create auth_totp ACL

This ACL allow users to generate TOTP Secret and change your password.

```shell
create_auth_acl
```

### Create ACL for a KV secret

This ACL allow users to read, write and delete secrets from path `staging/data/staging/*`

```shell
create_acl staging "staging/data/staging/*" read,write,delete
```

### Enable TOTP MFA

Enable the TOTP engine with name `totp` and issuer name `MyOrg`

```shell
create_mfa_totp totp MyOrg
```

### Create User with access on staging KV

Create user with random password and the policies `staging_read` and `default`

```shell
USERPWD="$(date | md5 | cut -b-20)"
create_user john "$USERPWD" "staging_read,default"
```

### Temporary user token

Create temporary user token to allow user set a new password and generate TOTP secret

```shell
#USERPWD random password generated in prior step. 
generate_user_token john "$USERPWD" 12h
```

### Enable MFA for user

Enforcement MFA login for user with TOTP

```shell
create_login_enforcement_entity totp john
```

### Disable MFA for user

Remove enforcement MFA login for user

```shell
create_login_enforcement_entity totp john delete
```

### Remove TOTP Secret for an user

```shell
destroy_totp_secret totp john
```

## As a user

Start a docker container to use a vault CLI.

```shell
docker run --rm -it --name vault-cli \
  -w /scripts \
  -e VAULT_ADDR="${VAULT_ADDR}" \
  hashicorp/vault:1.15 sh
```

From the container shell, run the following commands to update password and generate a TOTP secret

```shell
. <(wget -q -O- https://raw.githubusercontent.com/edimarlnx/secure-templates/main/dev/vault/mfa-with-username/user-func-utils.sh)
user_update_password USERNAME NEW_PASSWORD USER_TOKEN
user_generate_totp_secret METHOD_NAME USERNAME USER_TOKEN
```

For security reason, you can check the content of scripts loaded from this
step [here](https://raw.githubusercontent.com/edimarlnx/secure-templates/main/dev/vault/mfa-with-username/user-func-utils.sh)