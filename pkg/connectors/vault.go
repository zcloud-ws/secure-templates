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

func (v *VaultConnector) Secret(secretName, keyName string) any {
	kvSecret := v.kvSecrets[secretName]
	if kvSecret == nil {
		kvSec, err := v.client.KVv2(v.engineName).Get(context.Background(), fmt.Sprintf("%s/%s", v.ns, secretName))
		if err != nil {
			log.Fatalf("unable to read secret: %v", err)
			return keyName
		}
		v.kvSecrets[secretName] = kvSec
		kvSecret = kvSec
	}
	if keyName != "" {
		value, ok := kvSecret.Data[keyName].(string)
		if !ok {
			log.Fatalf("unable to load value for key %s", keyName)
			return keyName
		}
		return value
	}
	data := map[string]interface{}{}
	for k, vl := range kvSecret.Data {
		data[k] = vl
	}
	return data
}

func (v *VaultConnector) WriteKey(secretName, keyName, keyValue string) error {
	return v.WriteKeys(secretName, map[string]string{keyName: keyValue})
}

func (v *VaultConnector) WriteKeys(secretName string, keyValue map[string]string) error {
	secretPath := fmt.Sprintf("%s/%s", v.ns, secretName)
	data := map[string]interface{}{}
	for key, value := range keyValue {
		data[key] = value
	}
	_, err := v.client.KVv2(v.engineName).Patch(context.Background(), secretPath, data)
	if err != nil {
		log.Fatalf("unable to write secret: %v", err)
		return err
	}
	return nil
}

func (v *VaultConnector) Finalize() {

}
