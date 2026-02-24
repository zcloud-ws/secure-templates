package connectors

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/keymanagement"
	"github.com/oracle/oci-go-sdk/v65/secrets"
	"github.com/oracle/oci-go-sdk/v65/vault"
	"github.com/zcloud-ws/secure-templates/pkg/config"
	"github.com/zcloud-ws/secure-templates/pkg/envs"
	"github.com/zcloud-ws/secure-templates/pkg/helpers"
	"github.com/zcloud-ws/secure-templates/pkg/logging"
)

type OCIVaultConnector struct {
	Connector
	secretsClient                secrets.SecretsClient
	vaultsClient                 vault.VaultsClient
	kmsVaultClient               keymanagement.KmsVaultClient
	compartmentOCID              string
	keyOCID                      string
	defaultVaultOCID             string
	secretsCache                 map[string]string // "vaultName/secretName" → value
	vaultOCIDCache               map[string]string // vault name → OCID
	secretOCIDCache              map[string]string // "vaultOCID/secretName" → secret OCID
	secretIgnoreNotFoundKey      bool
	secretShowNameAsValueIfEmpty bool
	connectorType                config.SecretEngine
}

func (o *OCIVaultConnector) Init(secTplConfig config.SecureTemplateConfig) error {
	o.connectorType = secTplConfig.SecretEngine
	o.secretsCache = map[string]string{}
	o.vaultOCIDCache = map[string]string{}
	o.secretOCIDCache = map[string]string{}
	o.secretShowNameAsValueIfEmpty = secTplConfig.Options.SecretShowNameAsValueIfEmpty
	o.secretIgnoreNotFoundKey = secTplConfig.Options.SecretIgnoreNotFoundKey

	configFile := helpers.GetEnv(envs.OCIConfigFileEnv, secTplConfig.OCIVaultConfig.ConfigFile)
	if configFile == "" {
		configFile = "~/.oci/config"
	}
	profile := helpers.GetEnv(envs.OCIConfigProfileEnv, secTplConfig.OCIVaultConfig.Profile)
	if profile == "" {
		profile = "DEFAULT"
	}

	provider, err := common.ConfigurationProviderFromFileWithProfile(configFile, profile, "")
	if err != nil {
		return fmt.Errorf("unable to initialize OCI configuration provider: %v", err)
	}

	secretsClient, err := secrets.NewSecretsClientWithConfigurationProvider(provider)
	if err != nil {
		return fmt.Errorf("unable to initialize OCI Secrets client: %v", err)
	}
	o.secretsClient = secretsClient

	vaultsClient, err := vault.NewVaultsClientWithConfigurationProvider(provider)
	if err != nil {
		return fmt.Errorf("unable to initialize OCI Vaults client: %v", err)
	}
	o.vaultsClient = vaultsClient

	kmsVaultClient, err := keymanagement.NewKmsVaultClientWithConfigurationProvider(provider)
	if err != nil {
		return fmt.Errorf("unable to initialize OCI KMS Vault client: %v", err)
	}
	o.kmsVaultClient = kmsVaultClient

	o.defaultVaultOCID = helpers.GetEnv(envs.OCIVaultOCIDEnv, secTplConfig.OCIVaultConfig.VaultOCID)
	o.compartmentOCID = helpers.GetEnv(envs.OCICompartmentOCIDEnv, secTplConfig.OCIVaultConfig.CompartmentOCID)
	o.keyOCID = helpers.GetEnv(envs.OCIKeyOCIDEnv, secTplConfig.OCIVaultConfig.KeyOCID)

	return nil
}

// resolveVaultOCID resolves a vault display name to its OCID within the configured compartment.
func (o *OCIVaultConnector) resolveVaultOCID(vaultName string) (string, error) {
	if ocid, ok := o.vaultOCIDCache[vaultName]; ok {
		return ocid, nil
	}

	if o.compartmentOCID == "" {
		return "", errors.New("OCI Compartment OCID is required to resolve vault by name")
	}

	req := keymanagement.ListVaultsRequest{
		CompartmentId: common.String(o.compartmentOCID),
	}
	resp, err := o.kmsVaultClient.ListVaults(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("unable to list vaults: %v", err)
	}

	for _, v := range resp.Items {
		if v.DisplayName != nil && *v.DisplayName == vaultName &&
			v.LifecycleState == keymanagement.VaultSummaryLifecycleStateActive && v.Id != nil {
			o.vaultOCIDCache[vaultName] = *v.Id
			return *v.Id, nil
		}
	}

	return "", fmt.Errorf("vault '%s' not found in compartment", vaultName)
}

func (o *OCIVaultConnector) Secret(secretName, keyName string) any {
	var vaultName string
	var ociSecretName string

	if keyName == "" {
		// Single-arg call: {{ secret "name" }} → use default vault OCID
		if o.defaultVaultOCID == "" {
			logging.Log.Fatalf("default vault OCID is required for single-arg secret call\n")
			return secretName
		}
		vaultName = ""
		ociSecretName = secretName
	} else {
		// Two-arg call: {{ secret "vault-name" "secret-name" }}
		vaultName = secretName
		ociSecretName = keyName
	}

	cacheKey := vaultName + "/" + ociSecretName
	if cached, ok := o.secretsCache[cacheKey]; ok {
		return cached
	}

	// Resolve vault OCID
	var vaultOCID string
	var err error
	if vaultName == "" {
		vaultOCID = o.defaultVaultOCID
	} else {
		vaultOCID, err = o.resolveVaultOCID(vaultName)
		if err != nil {
			logging.Log.Fatalf("unable to resolve vault '%s': %v\n", vaultName, err)
			return ociSecretName
		}
	}

	req := secrets.GetSecretBundleByNameRequest{
		SecretName: common.String(ociSecretName),
		VaultId:    common.String(vaultOCID),
	}
	resp, err := o.secretsClient.GetSecretBundleByName(context.Background(), req)
	if err != nil {
		if !o.secretIgnoreNotFoundKey {
			logging.Log.Fatalf("unable to read OCI secret '%s': %v\n", ociSecretName, err)
		}
		logging.Log.Printf("unable to read OCI secret '%s': %v\n", ociSecretName, err)
		if o.secretShowNameAsValueIfEmpty {
			return ociSecretName
		}
		return ""
	}

	// Cache the secret OCID for potential write operations
	if resp.SecretId != nil {
		o.secretOCIDCache[vaultOCID+"/"+ociSecretName] = *resp.SecretId
	}

	content, ok := resp.SecretBundleContent.(secrets.Base64SecretBundleContentDetails)
	if !ok {
		logging.Log.Fatalf("unexpected secret content type for '%s'\n", ociSecretName)
		return ociSecretName
	}

	decoded, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		logging.Log.Fatalf("unable to decode secret content for '%s': %v\n", ociSecretName, err)
		return ociSecretName
	}

	value := string(decoded)
	o.secretsCache[cacheKey] = value
	return value
}

func (o *OCIVaultConnector) WriteKey(secretName, keyName, keyValue string) error {
	// secretName = vault name, keyName = secret name in OCI
	if o.compartmentOCID == "" {
		return errors.New("OCI Compartment OCID is required for write operations")
	}
	if o.keyOCID == "" {
		return errors.New("OCI Key OCID is required for write operations")
	}

	vaultOCID, err := o.resolveVaultOCID(secretName)
	if err != nil {
		return fmt.Errorf("unable to resolve vault '%s': %v", secretName, err)
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(keyValue))

	// Check if secret already exists
	secretOCID, exists, err := o.findSecretByName(vaultOCID, keyName)
	if err != nil {
		return fmt.Errorf("unable to check existing secret '%s': %v", keyName, err)
	}

	if exists {
		// Update existing secret with a new version
		req := vault.UpdateSecretRequest{
			SecretId: common.String(secretOCID),
			UpdateSecretDetails: vault.UpdateSecretDetails{
				SecretContent: vault.Base64SecretContentDetails{
					Content: common.String(encoded),
				},
			},
		}
		_, err = o.vaultsClient.UpdateSecret(context.Background(), req)
		if err != nil {
			return fmt.Errorf("unable to update OCI secret '%s': %v", keyName, err)
		}
	} else {
		// Create new secret
		req := vault.CreateSecretRequest{
			CreateSecretDetails: vault.CreateSecretDetails{
				CompartmentId: common.String(o.compartmentOCID),
				VaultId:       common.String(vaultOCID),
				KeyId:         common.String(o.keyOCID),
				SecretName:    common.String(keyName),
				SecretContent: vault.Base64SecretContentDetails{
					Content: common.String(encoded),
				},
			},
		}
		_, err = o.vaultsClient.CreateSecret(context.Background(), req)
		if err != nil {
			return fmt.Errorf("unable to create OCI secret '%s': %v", keyName, err)
		}
	}

	// Invalidate cache
	cacheKey := secretName + "/" + keyName
	delete(o.secretsCache, cacheKey)
	delete(o.secretOCIDCache, vaultOCID+"/"+keyName)

	return nil
}

func (o *OCIVaultConnector) WriteKeys(secretName string, keyValue map[string]string) error {
	for key, value := range keyValue {
		if err := o.WriteKey(secretName, key, value); err != nil {
			return err
		}
	}
	return nil
}

// findSecretByName searches for a secret by name in the specified vault.
func (o *OCIVaultConnector) findSecretByName(vaultOCID, secretName string) (string, bool, error) {
	cacheKey := vaultOCID + "/" + secretName
	if ocid, ok := o.secretOCIDCache[cacheKey]; ok {
		return ocid, true, nil
	}

	if o.compartmentOCID == "" {
		return "", false, errors.New("OCI Compartment OCID is required to list secrets")
	}

	req := vault.ListSecretsRequest{
		CompartmentId: common.String(o.compartmentOCID),
		VaultId:       common.String(vaultOCID),
		Name:          common.String(secretName),
	}
	resp, err := o.vaultsClient.ListSecrets(context.Background(), req)
	if err != nil {
		return "", false, err
	}

	for _, item := range resp.Items {
		if item.SecretName != nil && *item.SecretName == secretName {
			ocid := ""
			if item.Id != nil {
				ocid = *item.Id
				o.secretOCIDCache[cacheKey] = ocid
			}
			return ocid, true, nil
		}
	}

	return "", false, nil
}

func (o *OCIVaultConnector) Finalize() {
}

func (o *OCIVaultConnector) ConnectorType() config.SecretEngine {
	return o.connectorType
}
