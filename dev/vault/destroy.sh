#!/usr/bin/env bash
set -e

docker-compose -p vault down
docker volume rm vault_pg-data
docker volume rm vault_vault-data