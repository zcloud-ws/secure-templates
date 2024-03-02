#!/usr/bin/env sh

docker exec -it --env-file admin.env st_vault sh -c 'vault operator unseal $VAULT_UNSEAL_KEY'
