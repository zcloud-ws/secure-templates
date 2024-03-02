#!/usr/bin/env sh

docker exec -it --env-file admin.env st_vault sh -c 'vault secrets enable --path "kv/" kv-v2'
