#!/usr/bin/env sh

method_id_from_name() {
  vault list -format=table -detailed "/identity/mfa/method/totp" | grep "  ${1}" | cut -d" " -f1 | cat
}
entity_id_from_name() {
  vault read -field=id "/identity/entity/name/$1" | cat
}

user_update_password() {
  username="${1}"
  password="${2}"
  user_token="${3:-"$VAULT_USER_TOKEN"}"
  if [ "${username}" = "" ] || [ "${password}" = "" ] || [ "${user_token}" = "" ]; then
    echo "Required args 3 arguments: username password user_token"
    return 1
  fi
  VAULT_TOKEN="$user_token" vault write "auth/userpass/users/${username}/password" password="$password"
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
  VAULT_TOKEN="$user_token" vault write -field=url "/identity/mfa/method/totp/generate" \
    method_id="${METHOD_ID}" entity_id="${ENTITY_ID}" | cat
}


