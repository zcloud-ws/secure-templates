#!/usr/bin/env bash
set -e

echo Starting Database ...
docker-compose -p vault up -d pg

while ! docker exec -u postgres st_pg bash -l -c 'psql -d vault -c "select * from vault_ha_locks;" > /dev/null 2>&1'; do
  echo Waiting for database initialize ...
  sleep 1
done

echo Starting Vault ...
docker-compose -p vault up -d vault
