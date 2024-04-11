package render

import (
	"github.com/edimarlnx/secure-templates/pkg/connectors"
	"github.com/edimarlnx/secure-templates/pkg/helpers"
)

func RegisterSecret(connector connectors.Connector) func(args ...string) any {
	return func(args ...string) any {
		if len(args) == 1 {
			return connector.Secret(args[0], "")
		}
		return connector.Secret(args[0], args[1])
	}
}

func EnvVar(envName string) string {
	return helpers.GetEnv(envName, envName)
}
