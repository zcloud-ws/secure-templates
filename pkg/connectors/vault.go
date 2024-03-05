package connectors

import (
	"context"
	"errors"
	"fmt"
	"github.com/edimarlnx/secure-templates/pkg/config"
	"github.com/edimarlnx/secure-templates/pkg/helpers"
	vApi "github.com/hashicorp/vault/api"
	"log"
	"strings"
)

type VaultConnector struct {
	Connector
	client     *vApi.Client
	engineName string
	ns         string
	kvSecrets  map[string]*vApi.KVSecret
}

func (v *VaultConnector) Init(secTplConfig config.SecureTemplateConfig) error {
	cfg := vApi.DefaultConfig()
	cfg.Address = helpers.GetEnv("VAULT_ADDR", secTplConfig.VaultConfig.Address)
	client, err := vApi.NewClient(cfg)
	if err != nil {
		msg := fmt.Sprintf("unable to initialize Vault client: %v", err)
		return errors.New(msg)
	}
	token := helpers.GetEnv("VAULT_TOKEN", secTplConfig.VaultConfig.Token)
	if strings.TrimSpace(token) == "" {
		msg := "vault token is required"
		return errors.New(msg)
	}
	client.SetToken(token)
	v.client = client
	v.engineName = helpers.GetEnv("VAULT_SECRET_ENGINE", secTplConfig.VaultConfig.SecretEngine)
	if v.engineName == "" {
		v.engineName = "kv"
	}
	v.ns = helpers.GetEnv("VAULT_NS", secTplConfig.VaultConfig.Namespace)
	if v.ns == "" {
		v.ns = "dev"
	}
	v.kvSecrets = map[string]*vApi.KVSecret{}
	return nil
}

func (v *VaultConnector) Secret(secretName, keyName string) string {
	kvSecret := v.kvSecrets[secretName]
	if kvSecret == nil {
		kvSec, err := v.client.KVv2(v.engineName).Get(context.Background(), fmt.Sprintf("%s/%s", v.ns, secretName))
		if err != nil {
			log.Fatalf("unable to read secret: %v", err)
			return keyName
		}
		v.kvSecrets[keyName] = kvSec
		kvSecret = kvSec
	}

	value, ok := kvSecret.Data[keyName].(string)
	if !ok {
		log.Fatalf("unable to load value for key %s", keyName)
		return keyName
	}
	return value
}

func (v *VaultConnector) WriteKey(_, _, _ string) error {
	return errors.New("not implemented for Vault")
}

func (v *VaultConnector) Finalize() {

}
