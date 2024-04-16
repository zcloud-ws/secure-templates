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
	client                       *vApi.Client
	engineName                   string
	ns                           string
	kvSecrets                    map[string]*vApi.KVSecret
	secretIgnoreNotFoundKey      bool
	secretShowNameAsValueIfEmpty bool
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
	v.kvSecrets = map[string]*vApi.KVSecret{}
	v.secretShowNameAsValueIfEmpty = secTplConfig.Options.SecretShowNameAsValueIfEmpty
	v.secretIgnoreNotFoundKey = secTplConfig.Options.SecretIgnoreNotFoundKey
	return nil
}

func (v *VaultConnector) Secret(secretName, keyName string) any {
	kvSecret := v.kvSecrets[secretName]
	if kvSecret == nil {
		var mountPath string
		if v.ns != "" {
			mountPath = fmt.Sprintf("%s/%s", v.ns, secretName)
		} else {
			mountPath = secretName
		}
		kvSec, err := v.client.KVv2(v.engineName).Get(context.Background(), mountPath)
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
			if !v.secretIgnoreNotFoundKey {
				log.Fatalf("unable to load value for key %s", keyName)
			}
			log.Printf("unable to load value for key %s", keyName)
			if v.secretShowNameAsValueIfEmpty {
				return keyName
			}
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
	var secretPath string
	if v.ns != "" {
		secretPath = fmt.Sprintf("%s/%s", v.ns, secretName)
	} else {
		secretPath = secretName
	}
	data := map[string]interface{}{}
	for key, value := range keyValue {
		data[key] = value
	}
	secret, err := v.client.Logical().ReadWithContext(context.Background(), fmt.Sprintf("%s/data/%s", v.engineName, secretPath))
	if err != nil {
		log.Fatalf("unable to get secret: %v", err)
		return err
	}
	if secret == nil {
		_, err = v.client.KVv2(v.engineName).Put(context.Background(), secretPath, data)
	} else {
		_, err = v.client.KVv2(v.engineName).Patch(context.Background(), secretPath, data)
	}
	if err != nil {
		log.Fatalf("unable to write secret: %v", err)
		return err
	}
	return nil
}

func (v *VaultConnector) Finalize() {

}
