package connectors

import (
	"github.com/zcloud-ws/secure-templates/pkg/config"
	"github.com/zcloud-ws/secure-templates/pkg/logging"
)

type Connector interface {
	Init(secTplConfig config.SecureTemplateConfig) error
	Secret(secretName, keyName string) any
	WriteKey(secretName, keyName, keyValue string) error
	WriteKeys(secretName string, keyValue map[string]string) error
	Finalize()
	ConnectorType() config.SecretEngine
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
	case config.SecretEngineNo:
		connector = &NoConnector{}
	default:
		logging.Log.Fatalf("Connector not implemented: %s\n", secTplConfig.SecretEngine)
		return nil
	}
	err := connector.Init(secTplConfig)
	if err != nil {
		logging.Log.Fatalf("Error on init connector: %s\n", err.Error())
	}
	return connector
}
