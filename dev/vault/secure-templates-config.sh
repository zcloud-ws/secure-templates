#!/usr/bin/env sh

echo "Creating Secure Templates config for dev namespace."
DEV_TOKEN="$(./vault-dev-token.sh | sed 's/\\n//')"
cat <<EOF> "secure-templates-cfg.json"
{
  "secret_engine": "vault",
  "vault_config": {
    "address": "http://localhost:8200",
    "token": ${DEV_TOKEN},
    "secret_engine": "kv",
    "ns": "dev"
  }
}

EOF