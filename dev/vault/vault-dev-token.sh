#!/usr/bin/env sh
POLICY=${POLICY:-"dev_read"}
TTL=${TTL:-"30d"}
docker exec -it --env-file admin.env -e VAULT_FORMAT=json st_vault sh -c "vault token create -policy=${POLICY} -ttl ${TTL} -field=token | cat"
