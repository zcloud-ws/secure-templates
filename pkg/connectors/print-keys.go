package connectors

import (
	"errors"
	"github.com/zcloud-ws/secure-templates/pkg/config"
	"strings"
)

type PrintKeysConnector struct {
	Connector
	Keys          map[string]int
	connectorType config.SecretEngine
}

func (v *PrintKeysConnector) Init(secTplConfig config.SecureTemplateConfig) error {
	v.connectorType = secTplConfig.SecretEngine
	v.Keys = make(map[string]int)
	return nil
}

func (v *PrintKeysConnector) Secret(secretName, keyName string) any {
	var keys []string
	if secretName != "" {
		keys = append(keys, secretName)
	}
	if keyName != "" {
		keys = append(keys, keyName)
	}
	key := strings.Join(keys, ".")
	if v.Keys[key] == 0 {
		v.Keys[key] = 1
	} else {
		v.Keys[key] += 1
	}
	return key
}

func (v *PrintKeysConnector) Finalize() {

}

func (v *PrintKeysConnector) WriteKey(_, _, _ string) error {
	return errors.New("not implemented for Print Keys")
}

func (v *PrintKeysConnector) WriteKeys(_ string, _ map[string]string) error {
	return errors.New("not implemented for Print Keys")
}

func (v *PrintKeysConnector) ConnectorType() config.SecretEngine {
	return v.connectorType
}
