package connectors

import (
	"errors"
	"fmt"
	"github.com/edimarlnx/secure-templates/pkg/config"
)

type PrintKeysConnector struct {
	Connector
	Keys map[string]int
}

func (v *PrintKeysConnector) Init(_ config.SecureTemplateConfig) error {
	return nil
}

func (v *PrintKeysConnector) Secret(secretName, keyName string) any {
	key := fmt.Sprintf("%s.%s", secretName, keyName)
	if v.Keys[key] == 0 {
		v.Keys[key] = 1
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
