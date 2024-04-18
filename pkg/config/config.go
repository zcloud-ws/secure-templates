package config

import (
	"encoding/json"
	"io"
)

type SecretEngine string

const (
	SecretEngineVault     SecretEngine = "vault"
	SecretEngineLocalFile SecretEngine = "local-file"
	SecretEnginePrintKeys SecretEngine = "print-keys"
	//SecretEngineOnePassword SecretEngine = "one-password"
)

type SecureTemplateConfigOptions struct {
	SecretShowNameAsValueIfEmpty       bool   `json:"secretShowNameAsValueIfEmpty"`
	SecretIgnoreNotFoundKey            bool   `json:"secretIgnoreNotFoundKey"`
	EnvShowNameAsValueIfEmpty          bool   `json:"envShowNameAsValueIfEmpty"`
	EnvAllowAccessToSecureTemplateEnvs bool   `json:"envAllowAccessToSecureTemplateEnvs"`
	EnvRestrictedNameRegex             string `json:"envRestrictedNameRegex"`
}

type SecureTemplateConfig struct {
	SecretEngine SecretEngine `json:"secret_engine"`
	VaultConfig  VaultConfig  `json:"vault_config,omitempty"`
	//OnePasswordConfig OnePasswordConfig `json:"one_password_config,omitempty"`
	LocalFileConfig LocalFileConfig             `json:"local_file_config,omitempty"`
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

func (cfg *SecureTemplateConfig) Json(out io.Writer) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}
