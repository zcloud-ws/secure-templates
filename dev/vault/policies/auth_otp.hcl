path "/identity/mfa/method/totp/generate" {
  capabilities = ["create", "read", "update"]
}
path "auth/userpass/users/{{identity.entity.name}}/password" {
  capabilities = ["create", "update"]
}