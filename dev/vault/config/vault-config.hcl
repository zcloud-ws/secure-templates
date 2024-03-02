ui            = true
cluster_addr  = "http://localhost:8201"
api_addr      = "http://localhost:8200"
disable_mlock = true

storage "postgresql" {
  connection_url = "postgres://postgres:postgres@pg:5432/vault"
}

listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_disable = true
}
