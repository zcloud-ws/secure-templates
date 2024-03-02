#!/usr/bin/env sh

export INIT_DATA=
INIT_DATA="$(docker exec -it st_vault sh -c 'vault operator init -key-shares=1 -key-threshold=1 | sed -r "s/[[:cntrl:]]\[[0-9]{1,3}m//g" 2> /dev/null || echo ""')"

if [ "${INIT_DATA}" != "" ]; then
  echo "${INIT_DATA}" > keys
  V_KEY="$(grep "Unseal Key 1" keys | sed 's/Unseal Key 1: //')"
  V_ROOT_TOKEN="$(grep "Initial Root Token: " keys | sed 's/Initial Root Token: //')"
  cat <<EOF > "admin.env"
VAULT_UNSEAL_KEY=${V_KEY}
VAULT_TOKEN=${V_ROOT_TOKEN}
EOF

  ./vault-unseal.sh
  ./vault-enable-kv.sh
  ./vault-init-data.sh
  ./secure-templates-config.sh
fi
