#!/usr/bin/env sh

echo "Creating secret values for dev namespace"
docker exec -it --env-file admin.env st_vault sh -c 'vault kv put -mount=kv dev/core app_user="dev_user" app_passwd="2dabe3d7c66fb75f751202fdab19266b"'
docker exec -it --env-file admin.env st_vault sh -c 'vault kv put -mount=kv dev/client app_user="dev_client" app_passwd="6550a7838720e7904d72cae076ceea83"'

echo "Creating ACL policies for dev access"
docker exec -it --env-file admin.env --env ACL="$(cat policies/dev_read.hcl)" st_vault sh -c 'echo "$ACL" | vault policy write dev_read -'
docker exec -it --env-file admin.env --env ACL="$(cat policies/dev_write.hcl)" st_vault sh -c 'echo "$ACL" | vault policy write dev_write -'
