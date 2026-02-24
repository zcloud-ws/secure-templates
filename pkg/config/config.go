package config

import (
	"encoding/json"
	"io"
	"os"
	"strings"
)

type SecretEngine string

const (
	SecretEngineVault     SecretEngine = "vault"
	SecretEngineLocalFile SecretEngine = "local-file"
	SecretEnginePrintKeys SecretEngine = "print-keys"
	SecretEngineNo        SecretEngine = "no"
	SecretEngineOCIVault  SecretEngine = "oci-vault"
	//SecretEngineOnePassword SecretEngine = "one-password"
)

type SecureTemplateConfigOptions struct {
	SecretShowNameAsValueIfEmpty       bool   `json:"secretShowNameAsValueIfEmpty"`
	SecretIgnoreNotFoundKey            bool   `json:"secretIgnoreNotFoundKey"`
	EnvShowNameAsValueIfEmpty          bool   `json:"envShowNameAsValueIfEmpty"`
	EnvAllowAccessToSecureTemplateEnvs bool   `json:"envAllowAccessToSecureTemplateEnvs"`
	EnvRestrictedNameRegex             string `json:"envRestrictedNameRegex"`
	LeftDelim                          string `json:"leftDelim"`
	RightDelim                         string `json:"rightDelim"`
}

type SecureTemplateConfig struct {
	SecretEngine SecretEngine `json:"secret_engine"`
	VaultConfig  VaultConfig  `json:"vault_config,omitempty"`
	//OnePasswordConfig OnePasswordConfig `json:"one_password_config,omitempty"`
	LocalFileConfig LocalFileConfig             `json:"local_file_config,omitempty"`
	OCIVaultConfig  OCIVaultConfig              `json:"oci_vault_config,omitempty"`
	Options         SecureTemplateConfigOptions `json:"options"`
}

type VaultConfig struct {
	Address      string `json:"address"`
	Token        string `json:"token,omitempty"`
	SecretEngine string `json:"secret_engine,omitempty"`
	Namespace    string `json:"ns,omitempty"`
}

type OnePasswordConfig struct {
}

type LocalFileConfig struct {
	Filename   string `json:"filename"`
	EncPrivKey string `json:"enc_priv_key,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
}

type OCIVaultConfig struct {
	ConfigFile      string `json:"config_file,omitempty"`
	Profile         string `json:"profile,omitempty"`
	VaultOCID       string `json:"vault_ocid,omitempty"`
	CompartmentOCID string `json:"compartment_ocid,omitempty"`
	KeyOCID         string `json:"key_ocid,omitempty"`
}

func (cfg *SecureTemplateConfig) Json(out io.Writer) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}

func (cfg *SecureTemplateConfig) ExpandEnvVars() {
	cfg.VaultConfig.expandEnvVars()
	cfg.LocalFileConfig.expandEnvVars()
	cfg.OCIVaultConfig.expandEnvVars()
}

func (vCfg *VaultConfig) expandEnvVars() {
	vCfg.Address = expandEnvironmentVariables(vCfg.Address)
	vCfg.Token = expandEnvironmentVariables(vCfg.Token)
	vCfg.Namespace = expandEnvironmentVariables(vCfg.Namespace)
	vCfg.SecretEngine = expandEnvironmentVariables(vCfg.SecretEngine)
}

func (lCfg *LocalFileConfig) expandEnvVars() {
	lCfg.Filename = expandEnvironmentVariables(lCfg.Filename)
	lCfg.EncPrivKey = expandEnvironmentVariables(lCfg.EncPrivKey)
	lCfg.Passphrase = expandEnvironmentVariables(lCfg.Passphrase)
}

func (oCfg *OCIVaultConfig) expandEnvVars() {
	oCfg.ConfigFile = expandEnvironmentVariables(oCfg.ConfigFile)
	oCfg.Profile = expandEnvironmentVariables(oCfg.Profile)
	oCfg.VaultOCID = expandEnvironmentVariables(oCfg.VaultOCID)
	oCfg.CompartmentOCID = expandEnvironmentVariables(oCfg.CompartmentOCID)
	oCfg.KeyOCID = expandEnvironmentVariables(oCfg.KeyOCID)
}

func expandEnvironmentVariables(env string) string {
	if env == "" || !strings.Contains(env, "$") {
		return env
	}
	return os.ExpandEnv(env)
}
