package connectors

import (
	"github.com/edimarlnx/secure-templates/pkg/config"
	"log"
)

type Connector interface {
	Init(secTplConfig config.SecureTemplateConfig) error
	Secret(secretName, keyName string) string
	WriteKey(secretName, keyName, keyValue string) error
	WriteKeys(secretName string, keyValue map[string]string) error
	Finalize()
}

func NewConnector(secTplConfig config.SecureTemplateConfig) Connector {
	var connector Connector
	switch secTplConfig.SecretEngine {
	case config.SecretEngineVault:
		connector = &VaultConnector{}
	case config.SecretEngineLocalFile:
		connector = &LocalFileConnector{}
	case config.SecretEnginePrintKeys:
		connector = &PrintKeysConnector{}
	default:
		log.Fatalf("Connector not implemented: %s", secTplConfig.SecretEngine)
	}
	err := connector.Init(secTplConfig)
	if err != nil {
		log.Fatalf("Error on init connector: %s", err.Error())
	}
	return connector
}
