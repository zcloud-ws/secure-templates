#!/usr/bin/env sh

method_id_from_name() {
  vault list -format=table -detailed "/identity/mfa/method/totp" | grep "  ${1}" | cut -d" " -f1 | cat
}

entity_id_from_name() {
  vault read -field=id "/identity/entity/name/$1" | cat
}

enable_kv2() {
  kv_name="${1:-kv}"
  vault secrets enable --path "${kv_name}/" kv-v2
}

enable_auth_userpass() {
  vault auth enable userpass
}

create_auth_acl() {
  vault policy write "auth_totp" - <<EOF
path "/identity/mfa/method/totp/generate" {
    capabilities = ["create", "read", "update"]
}
path "auth/userpass/users/{{identity.entity.name}}/password" {
    capabilities = ["create", "update"]
}
path "identity/entity/name/{{identity.entity.name}}" {
    capabilities = ["read"]
}
EOF
}

create_acl() {
  acl_name=${1}
  secret_path=${2}
  permissions=${3:-"read"}
  if [ "${acl_name}" = "" ] || [ "${secret_path}" = "" ]; then
    echo "Required args 2 arguments:"
    echo "acl_name: name of policy"
    echo 'secret_path: ex.: "kv/data/dev/*" '
    echo 'permissions: read,write,delete (default read)" '
    return 1
  fi
  ttl=${2:-"30d"}
  if echo "$permissions" | grep -q read; then
    vault policy write "${acl_name}_read" - <<EOF
path "${secret_path}" {
   capabilities = ["read", "list"]
}
EOF
  fi
  if echo "$permissions" | grep -q write; then
    vault policy write "${acl_name}_write" - <<EOF
path "${secret_path}" {
   capabilities = ["create", "update", "patch"]
}
EOF
  fi
  if echo "$permissions" | grep -q delete; then
    vault policy write "${acl_name}_delete" - <<EOF
path "${secret_path}" {
   capabilities = ["delete"]
}
EOF
  fi

}

create_user() {
  username="${1}"
  password="${2}"
  policies="${3:-default},auth_totp"
  if [ "${username}" = "" ] || [ "${password}" = "" ] || [ "${policies}" = "" ]; then
    echo "Required args 3 arguments: username password policies (comma-delimited)"
    return 1
  fi
  vault write "/auth/userpass/users/${username}" password="$password"
  vault write "/identity/entity" name="$username" policies="${policies}"
  entity_id="$(entity_id_from_name "$username")"
  mount_accessor="$(vault read -field=accessor /sys/auth/userpass | cat)"
  echo "EntityID: $entity_id - Mount Accessor: $mount_accessor"
  vault write "/identity/entity-alias" name="$username" canonical_id="${entity_id}" mount_accessor="$mount_accessor"
}

create_mfa_totp() {
  method_name="${1:-"totp"}"
  issuer="${2:-"vault"}"
  vault write -field=method_id "/identity/mfa/method/totp" method_name="$method_name" issuer="${issuer}"
}

create_login_enforcement_entity() {
  method_name="${1}"
  entity_name="${2}"
  operation="${3:-write}"
  if [ "${method_name}" = "" ] || [ "${entity_name}" = "" ]; then
    echo "Required args 2 arguments: method_name entity_name operation(write|delete)"
    return 1
  fi
  entity_id="$(entity_id_from_name "$entity_name")"
  method_id="$(method_id_from_name "$method_name")"
  vault "${operation}" "/identity/mfa/login-enforcement/${method_name}_${entity_name}" \
    mfa_method_ids="$method_id"\
    identity_entity_ids="$entity_id"
}

destroy_totp_secret() {
  method_name="${1}"
  username="${2}"
  if [ "${method_name}" = "" ] || [ "${username}" = "" ]; then
      echo "Required args 2 arguments: method_name username"
      return 1
    fi
  method_id="$(method_id_from_name "$method_name")"
  entity_id="$(entity_id_from_name "$username")"
  vault write "/identity/mfa/method/totp/admin-destroy" method_id="$method_id" entity_id="${entity_id}" > /dev/null
}

create_totp_secret() {
  method_name="${1}"
  username="${2}"
  if [ "${method_name}" = "" ] || [ "${username}" = "" ]; then
      echo "Required args 2 arguments: method_name username"
      return 1
    fi
  method_id="$(method_id_from_name "$method_name")"
  entity_id="$(entity_id_from_name "$username")"
  vault write -field=url "/identity/mfa/method/totp/generate" method_id="$method_id" entity_id="${entity_id}" | cat
}

generate_user_token() {
  username="${1}"
  password="${2}"
  ttl=${ttl:-"30d"}
  if [ "${username}" = "" ] || [ "${password}" = "" ]; then
    echo "Required args 2 arguments: username password ttl (30d)"
    return 1
  fi
  vault login -method=userpass -token-only username="${username}" password="${password}" | cat
}

show_user_pwd_otp_help() {
  username="${1}"
  method_name="${2}"
  user_token="${3}"
  ttl="${4}"
  if [ "${username}" = "" ] || [ "${method_name}" = "" ] || [ "${user_token}" = "" ] || [ "${ttl}" = "" ]; then
    echo "Required args 4 arguments: username method_name user_token ttl"
    return 1
  fi
  METHOD_ID="$(method_id_from_name "${method_name}")"

cat <<EOL
# User setup instructions

\`\`\`
Username: $username
User token \(valid for $ttl\): $user_token
\`\`\`

## Update password and create TOTP secret using docker container:

### Start Vault CLI container:
\`\`\`
docker run --rm -it --name vault-cli \\
  -e VAULT_ADDR="${VAULT_ADDR}" \\
  hashicorp/vault:1.15 sh
\`\`\`

### Execute the following commands on opened container shell:

\`\`\`
. <(wget -q -O- https://raw.githubusercontent.com/edimarlnx/secure-templates/main/scripts/vault/user-func-utils.sh)

user_update_password $user_token $username NEW_PASSWORD

user_generate_totp_secret $user_token ${METHOD_ID} $username
\`\`\`

EOL

}


create_or_reset_user() {
  totp_name="${1}"
  username="${2}"
  user_pass="${3}"
  user_policies="${4}"
  user_token_setup_ttl="${5}"
  if [ "${totp_name}" = "" ] || [ "${username}" = "" ] || [ "${user_pass}" = "" ] || [ "${user_policies}" = "" ]  || [ "${user_token_setup_ttl}" = "" ]; then
    echo "Required args 5 arguments: totp_name username user_pass user_policies user_token_setup_ttl"
    return 1
  fi
  # Create User and password with access policies
  create_user "$username" "$user_pass" "$user_policies"

  # Delete Login enforcement on TOTP for entity
  create_login_enforcement_entity "$totp_name" "$username" delete

  destroy_totp_secret "$totp_name" "$username"

  # Create user token
  USER_TOKEN_TEMP="$(generate_user_token "$username" "$user_pass" "$user_token_setup_ttl")"

  # Create Login enforcement on TOTP for entity
  create_login_enforcement_entity "$totp_name" "$username"

  show_user_pwd_otp_help "$username" "$totp_name" "$USER_TOKEN_TEMP" "$user_token_setup_ttl"
}