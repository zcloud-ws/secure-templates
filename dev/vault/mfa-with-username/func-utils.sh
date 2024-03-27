#!/usr/bin/env sh

method_id_from_name() {
  vault list -format=table -detailed "/identity/mfa/method/totp" | grep "  ${1}" | cut -d" " -f1 | cat
}
entity_id_from_name() {
  vault read -field=id "/identity/entity/name/$1" | cat
}
entity_ids_in_group_name() {
  vault read -field=member_entity_ids "/identity/group/name/$1" | sed 's/\]\|\[//g'
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
# Allow tokens to look up their own properties
path "/identity/mfa/method/totp/generate" {
    capabilities = ["create", "read", "update"]
}
# Allow tokens to look up their own properties
path "auth/userpass/users/{{identity.entity.name}}/password" {
    capabilities = ["create", "update"]
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

user_update_password() {
  username="${1}"
  password="${2}"
  user_token="${3:-"$VAULT_USER_TOKEN"}"
  if [ "${username}" = "" ] || [ "${password}" = "" ] || [ "${user_token}" = "" ]; then
    echo "Required args 3 arguments: username password user_token"
    return 1
  fi
  vault write -header="X-Vault-Token=$user_token" "auth/userpass/users/${username}/password" password="$password"
}

user_generate_totp_secret() {
  method_name="${1}"
  username="${2}"
  user_token="${3:-"$VAULT_USER_TOKEN"}"
  if [ "${username}" = "" ] || [ "${password}" = "" ] || [ "${user_token}" = "" ]; then
    echo "Required args 3 arguments: method_name password user_token"
    return 1
  fi
  METHOD_ID="$(method_id_from_name "${method_name}")"
  ENTITY_ID="$(entity_id_from_name "${username}")"
  vault write -header="X-Vault-Token=$user_token" -field=url "/identity/mfa/method/totp/generate" \
    method_id="${METHOD_ID}" entity_id="${ENTITY_ID}" | cat
}


