#!/usr/bin/env sh

entity_id_from_name() {
  vault read -field=id "/identity/entity/name/$1" | cat
}

user_update_password() {
  user_token="${1:-"$VAULT_USER_TOKEN"}"
  username="${2}"
  password="${3}"
  if [ "${username}" = "" ] || [ "${password}" = "" ] || [ "${user_token}" = "" ]; then
    echo "Required args 3 arguments: user_token username password"
    return 1
  fi
  VAULT_TOKEN="$user_token" vault write "auth/userpass/users/${username}/password" password="$password"
}

user_generate_totp_secret() {
  user_token="${1:-"$VAULT_USER_TOKEN"}"
  method_id="${2}"
  username="${3}"
  if [ "${user_token}" = "" ] || [ "${method_id}" = "" ] || [ "${username}" = "" ]; then
    echo "Required args 3 arguments: user_token method_id username"
    return 1
  fi
  ENTITY_ID="$(VAULT_TOKEN="$user_token" entity_id_from_name "${username}")"
  VAULT_TOKEN="$user_token" vault write -field=url "/identity/mfa/method/totp/generate" \
    method_id="${method_id}" entity_id="${ENTITY_ID}" | cat
}


