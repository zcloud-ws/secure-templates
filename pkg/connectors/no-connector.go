package connectors

import (
	"errors"
	"github.com/zcloud-ws/secure-templates/pkg/config"
	"github.com/zcloud-ws/secure-templates/pkg/logging"
	"strings"
)

type NoConnector struct {
	Connector
	connectorType config.SecretEngine
}

func (v *NoConnector) Init(secTplConfig config.SecureTemplateConfig) error {
	v.connectorType = secTplConfig.SecretEngine
	return nil
}

func (v *NoConnector) Secret(secretName, keyName string) any {
	var keys []string
	if secretName != "" {
		keys = append(keys, secretName)
	}
	if keyName != "" {
		keys = append(keys, keyName)
	}
	key := strings.Join(keys, ".")
	logging.Log.Warnf("Not implemented for no-connector %s\n", key)
	return key
}

func (v *NoConnector) Finalize() {

}

func (v *NoConnector) WriteKey(_, _, _ string) error {
	return errors.New("not implemented for no-connector")
}

func (v *NoConnector) WriteKeys(_ string, _ map[string]string) error {
	return errors.New("not implemented for no-connector")
}

func (v *NoConnector) ConnectorType() config.SecretEngine {
	return v.connectorType
}
